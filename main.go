package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"reflect"
	"syscall"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
)

type Cell func(chan<- string)

var StatusBar = []Cell{
	MPD,
	Weather(myZip),
	Loadavg,
	Volume,
	Battery("BAT0"),
	Clock,
}

func Run(bar []Cell) <-chan string {
	out := make(chan string)
	go func() {
		var ts = make([]string, len(bar))
		var cases = make([]reflect.SelectCase, len(bar))
		for i, cell := range bar {
			ch := make(chan string)
			go cell(ch)
			cases[i] = reflect.SelectCase{
				Dir:  reflect.SelectRecv,
				Chan: reflect.ValueOf(ch),
			}
		}
		for {
			index, value, _ := reflect.Select(cases)
			text := value.Interface().(string)
			if text != ts[index] {
				ts[index] = text
				out <- FormatStatus(ts, " | ")
			}
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

	statc := Run(StatusBar)
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case status := <-statc:
			updateStatus(status)
		case <-sigc:
			updateStatus("")
			os.Exit(1)
		}
	}
}
