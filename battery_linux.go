package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"time"
)

func Battery(bat string) func(chan<- string) {
	return func(ch chan<- string) {
		for {
			b, err := ioutil.ReadFile("/sys/class/power_supply/" + bat +
				"/capacity")
			if err != nil {
				log.Println(err)
				return
			}
			capacity, err := strconv.Atoi(string(b[:len(b)-1]))
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
