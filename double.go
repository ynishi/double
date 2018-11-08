package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

func handlerFunc(sema chan struct{}, err chan error) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
		fmt.Printf("start handle\n")
		if len(sema) >= 1 {
			err <- errors.New("busy")
			fmt.Fprintf(w, "busy, cap:%d, len:%d", cap(sema), len(sema))
			cancel()
			return
		} else {
			// lock
			sema <- struct{}{}
			go second(ctx, sema, err)
			fmt.Fprintf(w, "sent, cap:%d, len:%d", cap(sema), len(sema))
			return
		}
	}
}

type Second struct {
	Sema    chan struct{}
	Handler http.HandlerFunc
}

func second(ctx context.Context, sema chan struct{}, err chan error) {
	fmt.Println("second job start", time.Now())
	go func(ctx context.Context) {
		fmt.Println("inner start", time.Now())
		for {
			select {
			case <-ctx.Done():
				fmt.Println("inner canceled", time.Now())
				return
			case <-time.After(1 * time.Second):
				fmt.Println("inner sec...", time.Now())
			}
		}
		fmt.Println("inner end", time.Now())
	}(ctx)
	for i := 0; i < 5; i++ {
		select {
		case <-ctx.Done():
			fmt.Println("canceled", ctx.Err(), time.Now())
			<-sema
			return
		case <-time.After(1 * time.Second):
			fmt.Println("1 sec...", time.Now())
		}
	}
	fmt.Println("second job end", time.Now())
	// unlock
	<-sema
}

func errFunc(errC chan error) {
	for err := range errC {
		fmt.Println("catch err:", err)
	}
}

var defaultSecond = &Second{}

func SetSema(s *Second) *Second {
	s.Sema = make(chan struct{}, 1)
	return s
}
func SetHandler(s *Second, errCh chan error) *Second {
	s.Handler = handlerFunc(s.Sema, errCh)
	return s
}

func main() {
	s := defaultSecond
	s = SetSema(s)

	errCh := make(chan error)
	go func() {
		for err := range errCh {
			fmt.Println("catch err:", err)
		}
	}()
	s = SetHandler(s, errCh)

	http.HandleFunc("/", s.Handler)
	http.ListenAndServe(":8080", nil)
}
