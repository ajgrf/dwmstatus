package main

import (
	"fmt"
	"time"
)

func Clock(ch chan<- string) {
	for {
		now := time.Now()
		ch <- fmt.Sprintf("%s %.1s %5s", now.Format("1/2"), now.Format("Mon"),
			now.Format("3:04"))
		offset := time.Duration(time.Now().UnixNano()) % time.Minute
		time.Sleep(time.Minute - offset)
	}
}
