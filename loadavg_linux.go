package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"
)

func Loadavg(ch chan<- string) {
	for {
		b, err := ioutil.ReadFile("/proc/loadavg")
		if err != nil {
			log.Println(err)
			return
		}

		fields := strings.Fields(string(b))

		load1, err := strconv.ParseFloat(fields[0], 64)
		if err != nil {
			log.Println(err)
			return
		}
		load5, err := strconv.ParseFloat(fields[1], 64)
		if err != nil {
			log.Println(err)
			return
		}
		load15, err := strconv.ParseFloat(fields[2], 64)
		if err != nil {
			log.Println(err)
			return
		}

		if load1 > 2.0 || load5 > 1.0 {
			ch <- fmt.Sprintf("%.2f %.2f %.2f", load1, load5, load15)
		} else {
			ch <- ""
		}

		time.Sleep(10 * time.Second)
	}
}
