package main

import (
	"time"
)

func Join(a []string, sep string) string {
	var first bool
	var b string
	for i := range a {
		if a[i] != "" {
			if first {
				b += sep
			}
			b += a[i]
			first = true
		}
	}
	return b
}

func Uniq(in <-chan string, out chan<- string) {
	defer close(out)
	var last string
	for s, ok := <-in; ok; s, ok = <-in {
		if s != last {
			out <- s
		}
		last = s
	}
}

func Timeout(ival time.Duration, in <-chan string, out chan<- string) {
	ch := make(chan string)
	defer close(ch)
	go Uniq(ch, out)
	for {
		select {
		case msg, ok := <-in:
			if ok {
				ch <- msg
			} else {
				break
			}
		case <-time.After(ival):
			ch <- ""
		}
	}
}
