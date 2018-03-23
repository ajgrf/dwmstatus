package main

import (
	"reflect"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
)

type Cell func(chan<- string)

func Status(cells ...Cell) {
	// init X connection & find root window
	x, err := xgb.NewConn()
	if err != nil {
		panic(err)
	}
	root := xproto.Setup(x).DefaultScreen(x).Root

	var ts = make([]string, len(cells))
	var cases []reflect.SelectCase
	for _, cell := range cells {
		ch := make(chan string)
		go cell(ch)
		cases = append(cases, reflect.SelectCase{Dir: reflect.SelectRecv,
			Chan: reflect.ValueOf(ch)})
	}
	for {
		index, value, _ := reflect.Select(cases)
		text := value.Interface().(string)
		ts[index] = text
		status := Join(ts, " | ")

		// set root window name with status text
		xproto.ChangeProperty(x, xproto.PropModeReplace, root, xproto.AtomWmName,
			xproto.AtomString, 8, uint32(len(status)), []byte(status))
	}
}

func Join(a []string, sep string) string {
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
	Status(Volume, Clock)
}
