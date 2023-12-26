package main

import (
	"log"
	"time"
)

func main() {
	parseFlags()
	getEnv()

	log.Printf(`Started with config:
  Url for collect metric: %s%s
  Report interval: %d
  Poll interval: %d
`, serverAddress, baseURL, reportInterval, pollInterval)

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
