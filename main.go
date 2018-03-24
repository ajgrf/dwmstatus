package main

import (
	"flag"
	"fmt"
	"log"
	"reflect"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
)

type StatusBar []Cell

type Cell func(chan<- string)

func (bar StatusBar) Run() <-chan string {
	out := make(chan string)
	go func() {
		var ts = make([]string, len(bar))
		var cases []reflect.SelectCase
		for _, cell := range bar {
			ch := make(chan string)
			go cell(ch)
			cases = append(cases, reflect.SelectCase{Dir: reflect.SelectRecv,
				Chan: reflect.ValueOf(ch)})
		}
		for {
			index, value, _ := reflect.Select(cases)
			text := value.Interface().(string)
			ts[index] = text
			out <- FormatStatus(ts, " | ")
		}
	}()
	return out
}

// FormatStatus concatenates the elements of a like strings.Join, except it
// treats empty strings as if they weren't there at all
func FormatStatus(a []string, sep string) string {
	var first bool
	var b string
	for i := range a {
		if a[i] != "" {
			if first {
				b += sep
			}
			b += a[i]
			first = true
		}
	}
	return b
}

func main() {
	noSetRoot := flag.Bool("n", false,
		"print to stdout instead of setting root window name")
	flag.Parse()

	var updateStatus func(string)
	if *noSetRoot {
		updateStatus = func(status string) {
			fmt.Println(status)
		}
	} else {
		// init X connection & find root window
		x, err := xgb.NewConn()
		if err != nil {
			log.Fatalln(err)
		}
		defer x.Close()
		root := xproto.Setup(x).DefaultScreen(x).Root

		updateStatus = func(status string) {
			// set root window name with status text
			xproto.ChangeProperty(x, xproto.PropModeReplace, root,
				xproto.AtomWmName, xproto.AtomString, 8, uint32(len(status)),
				[]byte(status))
		}
	}

	stats := StatusBar{Volume, Clock}.Run()
	for {
		updateStatus(<-stats)
	}
}
