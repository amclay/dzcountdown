package main

import (
	"fmt"
	"log"
	"time"
)

type timerFunc func(id string)

func startTimer(id string, callbackFunc timerFunc, timeout int) {
	stopTime := time.Now().Add(time.Duration(timeout) * time.Second)

	log.Printf("starting timer for location: %s", id)
	if a, ok := timers.Load(id); ok {
		timer := a.(*time.Timer)
		timer.Stop()
		timers.Delete(id)
	}
	timer := time.NewTimer(time.Duration(timeout) * time.Second)
	timers.Store(id, timer)

	go func() {
		for {
			select {
			case <-timer.C:
				callbackFunc(id)
				timers.Delete(id)
				return
			default:
				// delete if not loaded
				if _, ok := timers.Load(id); !ok {
					timers.Delete(id)
					return
				}
				time.Sleep(100 * time.Millisecond)
				durationLeft := time.Until(stopTime)
				minutes := int(durationLeft.Minutes())
				seconds := int(durationLeft.Seconds()) - (minutes * 60)
				updateText(id, fmt.Sprintf("%02d:%02d", minutes, seconds))
			}
		}
	}()
}
func stopTimer(id string) {
	log.Printf("stopping timer for location: %s", id)
	if a, ok := timers.Load(id); ok {
		timer := a.(*time.Timer)
		timer.Stop()
		timers.Delete(id)
	}
}
