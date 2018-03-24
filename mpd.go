package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fhs/gompd/mpd"
)

func MPD(ch chan<- string) {
	defer func() {
		ch <- ""
	}()

	var mpdHost, mpdPort, mpdPassword string
	mpdHost = os.Getenv("MPD_HOST")
	mpdPort = os.Getenv("MPD_PORT")
	if mpdHost == "" {
		mpdHost = "localhost"
	} else if strings.Contains(mpdHost, "@") {
		fields := strings.SplitN(mpdHost, "@", 2)
		mpdPassword = fields[0]
		mpdHost = fields[1]
	}
	if mpdPort == "" {
		mpdPort = "6600"
	}

	conn, err := mpd.DialAuthenticated("tcp", mpdHost+":"+mpdPort, mpdPassword)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	go func() {
		for conn.Ping() == nil {
			time.Sleep(30 * time.Second)
		}
	}()

	w, err := mpd.NewWatcher("tcp", mpdHost+":"+mpdPort, mpdPassword, "player")
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
