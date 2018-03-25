package main

import (
	"io/ioutil"
	"strconv"
)

func GetBattery(bat string) (int, error) {
	b, err := ioutil.ReadFile("/sys/class/power_supply/" + bat +
		"/capacity")
	if err != nil {
		return -1, err
	}
	capacity, err := strconv.Atoi(string(b[:len(b)-1]))
	if err != nil {
		return -1, err
	}

	return capacity, nil
}
