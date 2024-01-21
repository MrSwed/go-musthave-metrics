package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
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
		res    *http.Response
		v      interface{}
		metric = map[string]interface{}{"id": k, "type": t}
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
	urlStr := fmt.Sprintf("%s%s", serverAddress, baseURL)

	switch t {
	case gaugeType:
		metric["value"] = v
	case counterType:
		metric["delta"] = v
	default:
		err = fmt.Errorf("unknown metric type %s", t)
		return
	}
	var body []byte
	if body, err = json.Marshal(metric); err != nil {
		return
	}
	compressedBody := new(bytes.Buffer)

	zb := gzip.NewWriter(compressedBody)
	if _, err = zb.Write(body); err != nil {
		return
	}
	if err = zb.Close(); err != nil {
		return
	}
	var req *http.Request
	if req, err = http.NewRequest("POST", urlStr, compressedBody); err != nil {
		return
	}
	//req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	if res, err = http.DefaultClient.Do(req); err != nil {
		return
	}
	if err = res.Body.Close(); err != nil {
		return
	}
	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("post %s body %s: get StatusCode %d", urlStr, body, res.StatusCode)
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
