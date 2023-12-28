package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
)

type metricsCollects struct {
	memStats    runtime.MemStats
	pollCount   int64
	randomValue float64
	m           sync.RWMutex
}

func (m *metricsCollects) getMetrics() {
	m.m.Lock()
	defer m.m.Unlock()
	runtime.ReadMemStats(&m.memStats)
	m.pollCount++
	m.randomValue = rand.Float64()
}

func sendOneMetric(serverAddress, t, k string, v interface{}) (err error) {
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

func (m *metricsCollects) sendMetrics(serverAddress string) (err error) {
	if err = sendOneMetric(serverAddress, gaugeType, "Alloc", m.memStats.Alloc); err != nil {
		return
	}
	if err = sendOneMetric(serverAddress, gaugeType, "BuckHashSys", m.memStats.BuckHashSys); err != nil {
		return
	}
	if err = sendOneMetric(serverAddress, gaugeType, "Frees", m.memStats.Frees); err != nil {
		return
	}
	if err = sendOneMetric(serverAddress, gaugeType, "GCCPUFraction", m.memStats.GCCPUFraction); err != nil {
		return
	}
	if err = sendOneMetric(serverAddress, gaugeType, "GCSys", m.memStats.GCSys); err != nil {
		return
	}
	if err = sendOneMetric(serverAddress, gaugeType, "HeapAlloc", m.memStats.HeapAlloc); err != nil {
		return
	}
	if err = sendOneMetric(serverAddress, gaugeType, "HeapIdle", m.memStats.HeapIdle); err != nil {
		return
	}
	if err = sendOneMetric(serverAddress, gaugeType, "HeapInuse", m.memStats.HeapInuse); err != nil {
		return
	}
	if err = sendOneMetric(serverAddress, gaugeType, "HeapObjects", m.memStats.HeapObjects); err != nil {
		return
	}
	if err = sendOneMetric(serverAddress, gaugeType, "HeapReleased", m.memStats.HeapReleased); err != nil {
		return
	}
	if err = sendOneMetric(serverAddress, gaugeType, "HeapSys", m.memStats.HeapSys); err != nil {
		return
	}
	if err = sendOneMetric(serverAddress, gaugeType, "LastGC", m.memStats.LastGC); err != nil {
		return
	}
	if err = sendOneMetric(serverAddress, gaugeType, "Lookups", m.memStats.Lookups); err != nil {
		return
	}
	if err = sendOneMetric(serverAddress, gaugeType, "MCacheInuse", m.memStats.MCacheInuse); err != nil {
		return
	}
	if err = sendOneMetric(serverAddress, gaugeType, "MCacheSys", m.memStats.MCacheSys); err != nil {
		return
	}
	if err = sendOneMetric(serverAddress, gaugeType, "MSpanInuse", m.memStats.MSpanInuse); err != nil {
		return
	}
	if err = sendOneMetric(serverAddress, gaugeType, "MSpanSys", m.memStats.MSpanSys); err != nil {
		return
	}
	if err = sendOneMetric(serverAddress, gaugeType, "Mallocs", m.memStats.Mallocs); err != nil {
		return
	}
	if err = sendOneMetric(serverAddress, gaugeType, "NextGC", m.memStats.NextGC); err != nil {
		return
	}
	if err = sendOneMetric(serverAddress, gaugeType, "NumForcedGC", m.memStats.NumForcedGC); err != nil {
		return
	}
	if err = sendOneMetric(serverAddress, gaugeType, "NumGC", m.memStats.NumGC); err != nil {
		return
	}
	if err = sendOneMetric(serverAddress, gaugeType, "OtherSys", m.memStats.OtherSys); err != nil {
		return
	}
	if err = sendOneMetric(serverAddress, gaugeType, "PauseTotalNs", m.memStats.PauseTotalNs); err != nil {
		return
	}
	if err = sendOneMetric(serverAddress, gaugeType, "StackInuse", m.memStats.StackInuse); err != nil {
		return
	}
	if err = sendOneMetric(serverAddress, gaugeType, "StackSys", m.memStats.StackSys); err != nil {
		return
	}
	if err = sendOneMetric(serverAddress, gaugeType, "Sys", m.memStats.Sys); err != nil {
		return
	}
	if err = sendOneMetric(serverAddress, gaugeType, "TotalAlloc", m.memStats.TotalAlloc); err != nil {
		return
	}
	if err = sendOneMetric(serverAddress, counterType, "PollCount", m.pollCount); err != nil {
		return
	}
	if err = sendOneMetric(serverAddress, gaugeType, "RandomValue", m.randomValue); err != nil {
		return
	}

	return
}
