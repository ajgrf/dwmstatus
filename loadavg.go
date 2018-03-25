package main

import (
	"fmt"
	"log"
	"time"
)

func Loadavg(ch chan<- string) {
	for {
		loadavg, err := GetLoadavg()
		if err != nil {
			log.Println(err)
			return
		}

		if loadavg[0] > 2.0 || loadavg[1] > 1.0 {
			ch <- fmt.Sprintf("%.2f %.2f %.2f", loadavg[0], loadavg[1], loadavg[2])
		} else {
			ch <- ""
		}

		time.Sleep(10 * time.Second)
	}
}
