package main

import (
	"fmt"
	"log"
	"os"

	"github.com/fsnotify/fsnotify"
)

func Mail(maildir string) func(chan<- string) {
	return func(ch chan<- string) {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			log.Println(err)
			return
		}
		defer watcher.Close()

		err = watcher.Add(maildir + "/new")
		if err != nil {
			log.Println(err)
			return
		}

		for {
			select {
			case err := <-watcher.Errors:
				log.Println(err)
				ch <- ""
				return
			case <-watcher.Events:
				n, err := CountMsgs(maildir)
				if err != nil {
					log.Println(err)
					ch <- ""
					return
				}

				switch n {
				case 0:
					ch <- ""
				case 1:
					ch <- "1 new email"
				default:
					ch <- fmt.Sprint(n, " new emails")
				}
			}
		}

	}
}

func CountMsgs(maildir string) (int, error) {
	dir, err := os.Open(maildir + "/new")
	if err != nil {
		return -1, err
	}
	defer dir.Close()

	files, err := dir.Readdirnames(-1)
	if err != nil {
		return -1, err
	}

	return len(files), nil
}
