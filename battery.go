package main

import (
	"fmt"
	"log"
	"time"
)

func Battery(bat string) func(chan<- string) {
	return func(ch chan<- string) {
		for {
			capacity, err := GetBattery(bat)
			if err != nil {
				log.Println(err)
				return
			}

			if capacity < 98 {
				ch <- fmt.Sprintf("B:%3d%%", capacity)
			} else {
				ch <- ""
			}
			time.Sleep(60 * time.Second)
		}
	}
}
