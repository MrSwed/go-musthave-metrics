package main

import (
	"log"
	"os"
	"strconv"
	"time"
)

func getEnv() {
	addressEnv, reportIntervalEnv, pollIntervalEnv := os.Getenv("ADDRESS"), os.Getenv("REPORT_INTERVAL"), os.Getenv("POLL_INTERVAL")
	if addressEnv != "" {
		serverAddress = addressEnv
	}
	if reportIntervalEnv != "" {
		if v, err := strconv.ParseInt(reportIntervalEnv, 10, 64); err != nil {
			reportInterval = int(v)
		}
	}
	if pollIntervalEnv != "" {
		if v, err := strconv.ParseInt(pollIntervalEnv, 10, 64); err != nil {
			pollInterval = int(v)
		}
	}
}

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
