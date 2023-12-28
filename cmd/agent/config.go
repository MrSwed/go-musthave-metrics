package main

import (
	"flag"
	"os"
	"strconv"
	"strings"
)

const (
	baseURL     = "/update"
	gaugeType   = "gauge"
	counterType = "counter"
)

type config struct {
	serverAddress  string
	reportInterval int
	pollInterval   int
}

func (c *config) parseFlags() {
	flag.StringVar(&c.serverAddress, "a", "localhost:8080", "Provide the address of the metrics collection server")
	flag.IntVar(&c.reportInterval, "r", 10, "Provide the interval in seconds for send report metrics")
	flag.IntVar(&c.pollInterval, "p", 2, "Provide the interval in seconds for update metrics")
	flag.Parse()
}

func (c *config) getEnv() {
	addressEnv, reportIntervalEnv, pollIntervalEnv := os.Getenv("ADDRESS"), os.Getenv("REPORT_INTERVAL"), os.Getenv("POLL_INTERVAL")
	if addressEnv != "" {
		c.serverAddress = addressEnv
	}
	if reportIntervalEnv != "" {
		if v, err := strconv.Atoi(reportIntervalEnv); err == nil {
			c.reportInterval = v
		}
	}
	if pollIntervalEnv != "" {
		if v, err := strconv.Atoi(pollIntervalEnv); err == nil {
			c.pollInterval = v
		}
	}
}

func (c *config) getConfig() {
	c.parseFlags()
	c.getEnv()
	if !strings.HasPrefix(c.serverAddress, "http://") && !strings.HasPrefix(c.serverAddress, "https://") {
		c.serverAddress = "http://" + c.serverAddress
	}
}
