package main

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/fhs/gompd/mpd"
)

func MPD(ch chan<- string) {
	defer func() {
		ch <- ""
	}()

	mpdHost := os.Getenv("MPD_HOST")
	if mpdHost == "" {
		mpdHost = "localhost"
	}
	mpdPort := os.Getenv("MPD_PORT")
	if mpdPort == "" {
		mpdPort = "6600"
	}

	conn, err := mpd.Dial("tcp", mpdHost+":"+mpdPort)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	go func() {
		for err := conn.Ping(); err == nil; err = conn.Ping() {
			time.Sleep(30 * time.Second)
		}
	}()

	w, err := mpd.NewWatcher("tcp", mpdHost+":"+mpdPort, "", "player")
	if err != nil {
		log.Println(err)
		return
	}
	defer w.Close()

	go func() {
		for err := range w.Error {
			log.Println(err)
		}
	}()

	for {
		status, err := conn.Status()
		if err != nil {
			log.Println(err)
			return
		}

		if status["state"] == "play" {
			song, err := conn.CurrentSong()
			if err != nil {
				log.Println(err)
				return
			}

			s := song["Title"]
			if song["Artist"] != "" {
				s = song["Artist"] + " - " + s
			}
			if s == "" {
				s = filepath.Base(song["file"])
			}

			ch <- s
		} else {
			ch <- ""
		}

		<-w.Event
	}
}
