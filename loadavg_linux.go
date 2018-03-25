package main

import (
	"io/ioutil"
	"strconv"
	"strings"
)

func GetLoadavg() ([3]float64, error) {
	b, err := ioutil.ReadFile("/proc/loadavg")
	if err != nil {
		return [...]float64{-1, -1, -1}, err
	}

	fields := strings.Fields(string(b))

	load1, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return [...]float64{-1, -1, -1}, err
	}
	load5, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return [...]float64{-1, -1, -1}, err
	}
	load15, err := strconv.ParseFloat(fields[2], 64)
	if err != nil {
		return [...]float64{-1, -1, -1}, err
	}

	return [...]float64{load1, load5, load15}, nil
}
