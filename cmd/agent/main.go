package main

import (
	"log"
	"time"
)

func main() {
	var conf = new(config)

	conf.getConfig()

	log.Printf(`Started with config:
  Url for collect metric: %s%s
  Report interval: %d
  Poll interval: %d
`, conf.serverAddress, baseURL, conf.reportInterval, conf.pollInterval)

	lastSend := time.Now()
	m := new(metricsCollects)
	go func() {
		for {
			m.getMetrics()
			time.Sleep(time.Duration(conf.pollInterval) * time.Second)
		}
	}()

	for {
		if time.Now().After(lastSend.Add(time.Duration(conf.reportInterval) * time.Second)) {
			lastSend = time.Now()
			if err := m.sendMetrics(conf.serverAddress); err != nil {
				log.Print(err)
			} else {
				log.Print("metrics sent")
			}
		}
	}
}
