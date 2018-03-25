package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fhs/gompd/mpd"
)

// Formatted song displays will be truncated past this length
const maxSongLen = 100

func MPD(ch chan<- string) {
	// Clear MPD cell if this function exits
	defer func() {
		ch <- ""
	}()

	// Watch for MPD events
	w, err := mpd.NewWatcher(getMPDServerInfo())
	if err != nil {
		log.Println(err)
		return
	}
	defer w.Close()

	// Log watcher errors
	go func() {
		for err := range w.Error {
			log.Println(err)
		}
	}()

	// Only watch for 'player' events
	w.Subsystems("player")

	for ok := true; ok; _, ok = <-w.Event {
		song, err := formatCurrentSong()
		if err != nil {
			log.Println(err)
			return
		}

		ch <- song
	}
}

// getMPDServerInfo checks the environment and returns the information needed
// to dial an MPD server.
func getMPDServerInfo() (network, addr, password string) {
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

	return "tcp", mpdHost + ":" + mpdPort, mpdPassword
}

// formatCurrentSong returns the current song artist and title, or the empty
// string if no song is playing.
func formatCurrentSong() (string, error) {
	conn, err := mpd.DialAuthenticated(getMPDServerInfo())
	if err != nil {
		return "", err
	}
	defer conn.Close()

	status, err := conn.Status()
	if err != nil {
		return "", err
	}

	if status["state"] == "play" {
		song, err := conn.CurrentSong()
		if err != nil {
			return "", err
		}

		s := song["Title"]
		if song["Artist"] != "" {
			s = song["Artist"] + " - " + s
		}

		if s == "" {
			s = filepath.Base(song["file"])
		} else if len(s) > maxSongLen {
			s = s[:maxSongLen-3] + "..."
		}

		return s, nil
	}

	return "", nil
}
