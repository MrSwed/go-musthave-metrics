package main

import (
	"errors"
	"log"
	"time"
)

func main() {
	var conf = new(config)

	conf.Config()

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
			if errs := m.sendMetrics(conf.serverAddress, gaugeType, conf.gaugesList); errs != nil {
				log.Print(errors.Join(errs...))
			} else {
				log.Printf("%d Gauges metrics sent", len(conf.gaugesList))
			}
			if errs := m.sendMetrics(conf.serverAddress, counterType, conf.countersList); errs != nil {
				log.Print(errors.Join(errs...))
			} else {
				log.Printf("%d Counter metrics sent", len(conf.countersList))
			}
		}
	}
}
