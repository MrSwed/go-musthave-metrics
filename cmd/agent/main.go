package main

import (
	"log"
	"time"
)

func main() {
	parseFlags()

	lastSend := time.Now()
	for {
		getMetrics(m)
		if time.Now().After(lastSend.Add(time.Duration(reportInterval) * time.Second)) {
			lastSend = time.Now()
			if err := sendMetrics(m); err != nil {
				log.Print(err)
			} else {
				log.Print("metrics sent")
			}
		}
		time.Sleep(time.Duration(pollInterval) * time.Second)
	}
}
