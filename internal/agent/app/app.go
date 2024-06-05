package app

import (
	"context"
	"errors"
	"log"
	"net/url"
	"sync"
	"time"

	"go-musthave-metrics/internal/agent/config"
	"go-musthave-metrics/internal/agent/constant"
)

type BuildMetadata struct {
	Version string `json:"buildVersion"`
	Date    string `json:"buildDate"`
	Commit  string `json:"buildCommit"`
}

type app struct {
	wg    *sync.WaitGroup
	ctx   context.Context
	cfg   *config.Config
	m     *MetricsCollects
	build BuildMetadata
}

func NewApp(ctx context.Context, cfg *config.Config, metadata BuildMetadata) *app {
	return &app{
		ctx:   ctx,
		build: metadata,
		cfg:   cfg,
		m:     NewMetricsCollects(cfg),
		wg:    &sync.WaitGroup{},
	}
}

func (a *app) collectRuntime() {
	// collect runtime metrics
	log.Printf("daemon started: collect runtime metrics with interval %v", a.cfg.PollInterval)

	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		for {
			select {
			case <-time.After(time.Duration(a.cfg.PollInterval) * time.Second):
				log.Println("Collect runtime metrics")
				a.m.GetMetrics()
			case <-a.ctx.Done():
				log.Println("Runtime metrics collector is stopped")
				return
			}
		}
	}()
}

func (a *app) collectPSUtil() {
	// collect psutil metrics
	log.Printf("daemon started: collect psutil metrics with interval %v", a.cfg.PollInterval)

	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		for {
			select {
			case <-time.After(time.Duration(a.cfg.PollInterval) * time.Second):
				log.Println("Collect psutil metrics")
				if err := a.m.GetGopMetrics(); err != nil {
					log.Println("Error", err.Error())
				}
			case <-a.ctx.Done():
				log.Println("PSUtil metrics collector is stopped")
				return
			}
		}
	}()
}

func (a *app) sender() {
	a.wg.Add(1)
	log.Printf("daemon started: send metrics interval %v", a.cfg.ReportInterval)
	go func() {
		defer a.wg.Done()
		urlErr := &url.Error{}
		for {
			select {
			case <-time.After(time.Duration(a.cfg.ReportInterval) * time.Second):
				for i := 0; i <= len(config.Backoff); i++ {
					if n, err := a.m.SendMetrics(a.ctx); err != nil {
						if !errors.As(err, &urlErr) {
							log.Println(err)
							break
						}
						log.Printf("try %d: %s", i+1, err)
						if i < len(config.Backoff) {
							log.Printf("wait %d second before next try", config.Backoff[i]/time.Second)
							select {
							case <-a.ctx.Done():
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
			case <-a.ctx.Done():
				log.Println("Metrics sender is stopped")
				return
			}
		}
	}()

}

func (a *app) Run() {
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
		buildInfo(a.build.Version),
		buildInfo(a.build.Date),
		buildInfo(a.build.Commit),
		a.cfg.Address, constant.BaseURL, a.cfg.ReportInterval, a.cfg.PollInterval,
		a.cfg.RateLimit, a.cfg.SendSize, a.cfg.Key, a.cfg.CryptoKey, len(a.cfg.GaugesList)+len(a.cfg.CountersList))

	// collect runtime metrics
	a.collectRuntime()

	// collect psutil metrics
	a.collectPSUtil()

	// send metrics
	a.sender()
}

func (a *app) Stop() {
	a.wg.Wait()
	log.Println("Agent stopped")
}
