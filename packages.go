package main

import (
	"fmt"
	"os/exec"
	"regexp"
	"time"
)

func Packages(ch chan<- string) {
	for {
		switch n := NewPackages(); n {
		case 0:
		case 1:
			ch <- "1 new pkg"
		default:
			ch <- fmt.Sprintf("%v new pkgs", n)
		}
		time.Sleep(15 * time.Minute)
	}
}

func NewPackages() int {
	apt_get := exec.Command("/usr/bin/apt-get", "-q", "-y", "-s",
		"--ignore-hold", "--allow-unauthenticated", "dist-upgrade")
	output, err := apt_get.Output()
	if err != nil {
		panic(err)
	}
	numPkgs := len(regexp.MustCompile("\nInst ").FindAll(output, -1))
	return numPkgs
}
