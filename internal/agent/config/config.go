package config

import (
	"flag"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/MrSwed/go-musthave-metrics/internal/agent/constant"
)

var Backoff = [3]time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

type Config struct {
	ServerAddress  string
	ReportInterval int
	PollInterval   int
	Key            string
	MetricLists
}

type MetricLists struct {
	GaugesList   []string `type:"gauge"`
	CountersList []string `type:"counter"`
}

func (c *Config) parseFlags() {
	flag.StringVar(&c.ServerAddress, "a", "localhost:8080", "Provide the address of the metrics collection server")
	flag.IntVar(&c.ReportInterval, "r", 10, "Provide the interval in seconds for send report metrics")
	flag.IntVar(&c.PollInterval, "p", 2, "Provide the interval in seconds for update metrics")
	flag.StringVar(&c.Key, "k", "", "Provide the key")
	flag.Parse()
}

func (c *Config) getEnv() {
	addressEnv, reportIntervalEnv, pollIntervalEnv, key :=
		os.Getenv(constant.EnvNameServerAddress),
		os.Getenv(constant.EnvNameReportInterval),
		os.Getenv(constant.EnvNamePollInterval),
		os.Getenv(constant.EnvNameKey)
	if addressEnv != "" {
		c.ServerAddress = addressEnv
	}
	if reportIntervalEnv != "" {
		if v, err := strconv.Atoi(reportIntervalEnv); err == nil {
			c.ReportInterval = v
		}
	}
	if pollIntervalEnv != "" {
		if v, err := strconv.Atoi(pollIntervalEnv); err == nil {
			c.PollInterval = v
		}
	}
	if key != "" {
		c.Key = key
	}
}

func (c *Config) setGaugesList(m ...string) {
	if len(m) > 0 {
		c.GaugesList = m
	} else {
		c.GaugesList = []string{
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
			"TotalMemory",
			"FreeMemory",
			"CPUutilization",
		}
	}
}

func (c *Config) setCountersList(m ...string) {
	if len(m) > 0 {
		c.CountersList = m
	} else {
		c.CountersList = []string{
			"PollCount",
		}
	}
}

func (c *Config) Config() {
	c.parseFlags()
	c.getEnv()
	if !strings.HasPrefix(c.ServerAddress, "http://") && !strings.HasPrefix(c.ServerAddress, "https://") {
		c.ServerAddress = "http://" + c.ServerAddress
	}
	// metric list can be set later from args or env
	// move this to the appropriate functions
	c.setGaugesList()
	c.setCountersList()
}
