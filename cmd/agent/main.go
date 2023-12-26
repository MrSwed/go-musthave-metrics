package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"time"
)

const (
	baseURL     = "/update"
	gaugeType   = "gauge"
	counterType = "counter"
)

type metricsCollects struct {
	memStats    runtime.MemStats
	pollCount   int64
	randomValue float64
}

func getMetrics(m *metricsCollects) {
	runtime.ReadMemStats(&m.memStats)
	m.pollCount++
	m.randomValue = rand.Float64()
}

func sendOneMetric(t, k string, v interface{}) (err error) {
	var res *http.Response
	urlStr := fmt.Sprintf("%s%s/%s/%s/%v", serverAddress, baseURL, t, k, v)
	if res, err = http.Post(urlStr, "text/plain", nil); err != nil {
		return
	}
	defer func() { err = res.Body.Close() }()
	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("%v", res.StatusCode)
	}

	return
}

func sendMetrics(m *metricsCollects) (err error) {
	if err = sendOneMetric(gaugeType, "Alloc", m.memStats.Alloc); err != nil {
		return
	}
	if err = sendOneMetric(gaugeType, "BuckHashSys", m.memStats.BuckHashSys); err != nil {
		return
	}
	if err = sendOneMetric(gaugeType, "Frees", m.memStats.Frees); err != nil {
		return
	}
	if err = sendOneMetric(gaugeType, "GCCPUFraction", m.memStats.GCCPUFraction); err != nil {
		return
	}
	if err = sendOneMetric(gaugeType, "GCSys", m.memStats.GCSys); err != nil {
		return
	}
	if err = sendOneMetric(gaugeType, "HeapAlloc", m.memStats.HeapAlloc); err != nil {
		return
	}
	if err = sendOneMetric(gaugeType, "HeapIdle", m.memStats.HeapIdle); err != nil {
		return
	}
	if err = sendOneMetric(gaugeType, "HeapInuse", m.memStats.HeapInuse); err != nil {
		return
	}
	if err = sendOneMetric(gaugeType, "HeapObjects", m.memStats.HeapObjects); err != nil {
		return
	}
	if err = sendOneMetric(gaugeType, "HeapReleased", m.memStats.HeapReleased); err != nil {
		return
	}
	if err = sendOneMetric(gaugeType, "HeapSys", m.memStats.HeapSys); err != nil {
		return
	}
	if err = sendOneMetric(gaugeType, "LastGC", m.memStats.LastGC); err != nil {
		return
	}
	if err = sendOneMetric(gaugeType, "Lookups", m.memStats.Lookups); err != nil {
		return
	}
	if err = sendOneMetric(gaugeType, "MCacheInuse", m.memStats.MCacheInuse); err != nil {
		return
	}
	if err = sendOneMetric(gaugeType, "MCacheSys", m.memStats.MCacheSys); err != nil {
		return
	}
	if err = sendOneMetric(gaugeType, "MSpanInuse", m.memStats.MSpanInuse); err != nil {
		return
	}
	if err = sendOneMetric(gaugeType, "MSpanSys", m.memStats.MSpanSys); err != nil {
		return
	}
	if err = sendOneMetric(gaugeType, "Mallocs", m.memStats.Mallocs); err != nil {
		return
	}
	if err = sendOneMetric(gaugeType, "NextGC", m.memStats.NextGC); err != nil {
		return
	}
	if err = sendOneMetric(gaugeType, "NumForcedGC", m.memStats.NumForcedGC); err != nil {
		return
	}
	if err = sendOneMetric(gaugeType, "NumGC", m.memStats.NumGC); err != nil {
		return
	}
	if err = sendOneMetric(gaugeType, "OtherSys", m.memStats.OtherSys); err != nil {
		return
	}
	if err = sendOneMetric(gaugeType, "PauseTotalNs", m.memStats.PauseTotalNs); err != nil {
		return
	}
	if err = sendOneMetric(gaugeType, "StackInuse", m.memStats.StackInuse); err != nil {
		return
	}
	if err = sendOneMetric(gaugeType, "StackSys", m.memStats.StackSys); err != nil {
		return
	}
	if err = sendOneMetric(gaugeType, "Sys", m.memStats.Sys); err != nil {
		return
	}
	if err = sendOneMetric(gaugeType, "TotalAlloc", m.memStats.TotalAlloc); err != nil {
		return
	}
	if err = sendOneMetric(counterType, "PollCount", m.pollCount); err != nil {
		return
	}
	if err = sendOneMetric(gaugeType, "RandomValue", m.randomValue); err != nil {
		return
	}

	return
}

func main() {
	parseFlags()

	lastSend := time.Now()
	m := new(metricsCollects)
	for {
		getMetrics(m)
		if time.Now().After(lastSend.Add(time.Duration(reportInterval) * time.Second)) {
			lastSend = time.Now()
			if err := sendMetrics(m); err != nil {
				log.Print(err)
			} else {
				log.Print("metrics sent")
			}
		}
		time.Sleep(time.Duration(pollInterval) * time.Second)
	}
}
