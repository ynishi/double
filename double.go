package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

func handlerFunc (sema chan struct{}, err chan error) func (w http.ResponseWriter, r *http.Request) {
   return func (w http.ResponseWriter, r *http.Request) {
	   fmt.Printf("start handle\n")
	   if len(sema) == 1 {
	   	    err <- errors.New("busy")
		   fmt.Fprintf(w, "busy, cap:%d, len:%d", cap(sema), len(sema))
		   return
	   } else {
	   	   // lock
		   sema <- struct{}{}
		   go second(sema, err)
		   fmt.Fprintf(w, "sent, cap:%d, len:%d", cap(sema), len(sema))
		   return
	   }
   }
}

type Second struct {
	Sema chan struct{}
	Err chan error
}

func second(sema chan struct{}, err chan error) {
		fmt.Println("second job start", time.Now())
		time.Sleep(10 * time.Second)
		fmt.Println("second job end", time.Now())
		// unlock
		<- sema
}

func errFunc(errC chan error) {
	for err := range errC {
		fmt.Println("catch err:", err)
	}
}

func main() {
	s := &Second{}
	s.Sema = make(chan struct{}, 1)
	s.Err = make(chan error)

    go errFunc(s.Err)

	handler := handlerFunc(s.Sema, s.Err)
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}