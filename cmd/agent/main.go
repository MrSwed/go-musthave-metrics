package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	var (
		wg   sync.WaitGroup
		conf = new(config)
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	conf.Config()

	log.Printf(`Started with config:
  Url for collect metric: %s%s
  Report interval: %d
  Poll interval: %d
`, conf.serverAddress, baseURL, conf.reportInterval, conf.pollInterval)

	m := new(metricsCollects)

	// collect metrics
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-time.After(time.Duration(conf.pollInterval) * time.Second):
				log.Println("Collect metrics")
				m.getMetrics()
			case <-ctx.Done():
				log.Println("Metrics collector is stopped")
				return
			}
		}
	}()

	// send metrics
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-time.After(time.Duration(conf.reportInterval) * time.Second):
				if errs := m.sendMetrics(conf.serverAddress, conf.metricLists); errs != nil {
					log.Println(errors.Join(errs...))
				} else {
					log.Printf("%d metrics sent", len(conf.GaugesList)+len(conf.CountersList))
				}
			case <-ctx.Done():
				log.Println("Metrics sender is stopped")
				return
			}
		}
	}()
	wg.Wait()

	log.Println("Agent stopped")
}
