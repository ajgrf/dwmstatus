package main

import (
	"fmt"
	"log"
	"time"

	owm "github.com/briandowns/openweathermap"
)

func Weather(zip int) func(chan<- string) {
	return func(ch chan<- string) {
		for {
			w, err := owm.NewCurrent("F", "en", myOWMKey)
			if err != nil {
				log.Println(err)
				return
			}

			w.CurrentByZip(zip, "us")
			ch <- fmt.Sprintf("%.0f°F %v", w.Main.Temp, w.Weather[0].Description)

			time.Sleep(15 * time.Minute)
		}
	}
}
