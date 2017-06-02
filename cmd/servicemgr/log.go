package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

var debug int = 1

func LogInit() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGUSR1)
	go func() {
		for {
			select {
			case s := <-c:
				log.Printf("get signal %v, change debug from %d to %d\n", s, debug, debug^1)
				debug ^= 1
			}
		}
	}()
}

func Log(format string, v ...interface{}) {
	if debug > 0 {
		log.Printf(format, v...)
	}
}
