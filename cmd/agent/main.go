package main

import (
	"context"
	"errors"
	"log"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/MrSwed/go-musthave-metrics/internal/agent/app"
	"github.com/MrSwed/go-musthave-metrics/internal/agent/config"
	"github.com/MrSwed/go-musthave-metrics/internal/agent/constant"
)

var buildVersion string
var buildDate string
var buildCommit string

func buildInfo(s string) string {
	if s == "" {
		return "N/A"
	}
	return s
}

func main() {
	var (
		wg   sync.WaitGroup
		conf = config.NewConfig()
	)

	if err := conf.Init(); err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf(`Started with build info:
  BuildVersion: %s
  BuildDate: %s
  BuildCommit: %s
With config:
  Url for collect metric: %s%s
  Report interval: %d
  Poll interval: %d
  Rate limit: %d
  Number of metrics at once: %d
  Key: %s
  CryptoKey: %s
  Metric names count: %d
`,
		buildInfo(buildVersion),
		buildInfo(buildDate),
		buildInfo(buildCommit),
		conf.ServerAddress, constant.BaseURL, conf.ReportInterval, conf.PollInterval,
		conf.RateLimit, conf.SendSize, conf.Key, conf.CryptoKey, len(conf.GaugesList)+len(conf.CountersList))

	m := app.NewMetricsCollects(conf)

	// collect runtime metrics
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-time.After(time.Duration(conf.PollInterval) * time.Second):
				log.Println("Collect runtime metrics")
				m.GetMetrics()
			case <-ctx.Done():
				log.Println("Runtime metrics collector is stopped")
				return
			}
		}
	}()

	// collect psutil metrics
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-time.After(time.Duration(conf.PollInterval) * time.Second):
				log.Println("Collect psutil metrics")
				if err := m.GetGopMetrics(); err != nil {
					log.Println("Error", err.Error())
				}
			case <-ctx.Done():
				log.Println("PSUtil metrics collector is stopped")
				return
			}
		}
	}()

	// send metrics
	wg.Add(1)
	go func() {
		defer wg.Done()
		urlErr := &url.Error{}
		for {
			select {
			case <-time.After(time.Duration(conf.ReportInterval) * time.Second):
				for i := 0; i <= len(config.Backoff); i++ {
					if n, err := m.SendMetrics(ctx); err != nil {
						if !errors.As(err, &urlErr) {
							log.Println(err)
							break
						}
						log.Printf("try %d: %s", i+1, err)
						if i < len(config.Backoff) {
							log.Printf("wait %d second before next try", config.Backoff[i]/time.Second)
							select {
							case <-ctx.Done():
								log.Print("ctx done, do not try more")
								return
							case <-time.After(config.Backoff[i]):
							}
						}
					} else {
						log.Printf("%d metrics sent", n)
						break
					}
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
