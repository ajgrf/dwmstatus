package main

import (
	"flag"
	"fmt"
	"reflect"
)

type Cell func(chan<- string)

func Status(sep string, cells ...Cell) {
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
		fmt.Println(Join(ts, sep))
	}
}

func main() {
	sep := flag.String("sep", " | ", "separator string to be placed between elements")
	flag.Parse()
	Status(*sep, Volume, Clock)
}
