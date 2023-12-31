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
	gaugesList     []string
	countersList   []string
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

func (c *config) setGaugesList(m ...string) {
	if len(m) > 0 {
		c.gaugesList = m
	} else {
		c.gaugesList = []string{
			"Alloc",
			"BuckHashSys",
			"Frees",
			"GCCPUFraction",
			"GCSys",
			"HeapAlloc",
			"HeapIdle",
			"HeapInuse",
			"HeapObjects",
			"HeapReleased",
			"HeapSys",
			"LastGC",
			"Lookups",
			"MCacheInuse",
			"MCacheSys",
			"MSpanInuse",
			"MSpanSys",
			"Mallocs",
			"NextGC",
			"NumForcedGC",
			"NumGC",
			"OtherSys",
			"PauseTotalNs",
			"StackInuse",
			"StackSys",
			"Sys",
			"TotalAlloc",
			"RandomValue",
		}
	}
}

func (c *config) setCountersList(m ...string) {
	if len(m) > 0 {
		c.countersList = m
	} else {
		c.countersList = []string{
			"PollCount",
		}
	}
}

func (c *config) Config() {
	c.parseFlags()
	c.getEnv()
	if !strings.HasPrefix(c.serverAddress, "http://") && !strings.HasPrefix(c.serverAddress, "https://") {
		c.serverAddress = "http://" + c.serverAddress
	}
	// metric list can be set later from args or env
	// move this to the appropriate functions
	c.setGaugesList()
	c.setCountersList()
}
