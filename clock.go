package main

import (
	"time"
)

func Clock(ch chan<- string) {
	for {
		ch <- time.Now().Format("2006-01-02 15:04")
		offset := time.Duration(time.Now().UnixNano()) % time.Minute
		time.Sleep(time.Minute - offset)
	}
}
