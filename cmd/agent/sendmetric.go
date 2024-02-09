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
		res       *http.Response
		v         interface{}
		oneMetric = newMetric(k, t)
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

	if err = oneMetric.set(v); err != nil {
		return
	}
	var body []byte
	if body, err = json.Marshal(oneMetric); err != nil {
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

func (m *metricsCollects) sendMetrics(serverAddress string, lists metricLists) (errs []error) {
	lRefVal := reflect.ValueOf(lists)
	lRefType := reflect.TypeOf(lists)
	var mType string
	for i := 0; i < lRefVal.NumField(); i++ {
		if mType = lRefType.Field(i).Tag.Get("type"); mType == "" {
			continue
		}
		lItemRef := reflect.Indirect(lRefVal.Field(i))
		if !lItemRef.IsValid() {
			continue
		}
		if list, ok := lItemRef.Interface().([]string); ok {
			for _, mName := range list {
				if err := m.sendOneMetric(serverAddress, mType, mName); err != nil {
					errs = append(errs, err)
				}
			}
		}
	}
	return
}
