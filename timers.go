package main

import (
	"log"
	"time"
)

func startTimer(id string, timeout int) {
	log.Printf("starting timer for location: %s", id)

	stopTime := time.Now().Add(time.Duration(timeout) * time.Second)

	timers.Store(id, stopTime)
}
