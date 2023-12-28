package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"reflect"
	"runtime"
	"sync"
)

type metricsCollects struct {
	runtime.MemStats
	PollCount   int64
	RandomValue float64
	m           sync.RWMutex
}

func (m *metricsCollects) getMetrics() {
	m.m.Lock()
	defer m.m.Unlock()
	runtime.ReadMemStats(&m.MemStats)
	m.PollCount++
	m.RandomValue = rand.Float64()
}

func (m *metricsCollects) sendOneMetric(serverAddress, t, k string) (err error) {
	var (
		res *http.Response
		v   interface{}
	)
	dVal := reflect.Indirect(reflect.ValueOf(m))
	if refV := dVal.FieldByName(k); refV.IsValid() {
		m.m.RLock()
		v = refV.Interface()
		m.m.RUnlock()
	} else {
		err = fmt.Errorf("unknown metric name %s", k)
		return
	}
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

func (m *metricsCollects) sendMetrics(serverAddress, metricType string, list []string) (errs []error) {
	for _, mName := range list {
		if err := m.sendOneMetric(serverAddress, metricType, mName); err != nil {
			errs = append(errs, err)
		}
	}

	return
}
