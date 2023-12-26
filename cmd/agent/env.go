package main

import (
	"os"
	"strconv"
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
