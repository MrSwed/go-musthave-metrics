package main

import (
	"flag"
)

var (
	serverAddress  string
	reportInterval int
	pollInterval   int
)

func parseFlags() {
	flag.StringVar(&serverAddress, "a", "localhost:8080", "Provide the address of the metrics collection server")
	flag.IntVar(&reportInterval, "r", 10, "Provide the interval in seconds for send report metrics")
	flag.IntVar(&pollInterval, "p", 2, "Provide the interval in seconds for update metrics")
	flag.Parse()
}
