package app

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/MrSwed/go-musthave-metrics/internal/agent/config"
	"github.com/MrSwed/go-musthave-metrics/internal/agent/constant"
	"io"
	"math/rand"
	"net/http"
	"reflect"
	"runtime"
	"sync"
)

type MetricsCollects struct {
	runtime.MemStats
	PollCount   int64
	RandomValue float64
	m           sync.RWMutex
	c           *config.Config
}

func NewMetricsCollects(c *config.Config) *MetricsCollects {
	return &MetricsCollects{
		c: c,
	}
}

func (m *MetricsCollects) GetMetrics() {
	m.m.Lock()
	defer m.m.Unlock()
	runtime.ReadMemStats(&m.MemStats)
	m.PollCount++
	m.RandomValue = rand.Float64()
}

func (m *MetricsCollects) SendMetrics(serverAddress string, lists config.MetricLists) (err error) {
	var (
		metrics []*Metric
		er      error
	)

	mRefVal := reflect.Indirect(reflect.ValueOf(m))
	urlStr := serverAddress + constant.BaseURL

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
				var (
					v interface{}
				)
				if refV := mRefVal.FieldByName(mName); refV.IsValid() {
					m.m.RLock()
					v = refV.Interface()
					m.m.RUnlock()
				} else {
					err = errors.Join(err, fmt.Errorf("unknown metric name %s", mName))
					continue
				}
				oneMetric := NewMetric(mName, mType)
				if er = oneMetric.Set(v); er != nil {
					err = errors.Join(err, er)
					continue
				}
				metrics = append(metrics, oneMetric)
			}
		}
	}

	var body []byte
	if body, er = json.Marshal(metrics); er != nil {
		err = errors.Join(err, er)
		return
	}
	compressedBody := new(bytes.Buffer)

	zb := gzip.NewWriter(compressedBody)
	if _, er = zb.Write(body); er != nil {
		err = errors.Join(err, er)
		return
	}

	if er = zb.Close(); er != nil {
		err = errors.Join(err, er)
		return
	}
	var req *http.Request
	if req, er = http.NewRequest("POST", urlStr, compressedBody); er != nil {
		err = errors.Join(err, er)
		return
	}
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	if m.c != nil && m.c.Key != "" {

		h := hmac.New(sha256.New, []byte(m.c.Key))
		if _, err = h.Write(compressedBody.Bytes()); err != nil {
			err = errors.Join(err, er)
			return
		}
		req.Header.Set("HashSHA256", hex.EncodeToString(h.Sum(nil)))
	}

	var res *http.Response
	if res, er = http.DefaultClient.Do(req); er != nil {
		err = errors.Join(err, er)
		return
	}
	var resultBody []byte
	if resultBody, er = io.ReadAll(res.Body); er != nil && !errors.Is(er, io.EOF) {
		err = errors.Join(err, er)
		return
	}
	if er = res.Body.Close(); er != nil {
		err = errors.Join(err, er)
		return
	}
	if res.StatusCode != http.StatusOK {
		err = errors.Join(err, fmt.Errorf("post to %s with body: %s. Get: statusCode: %d;  answer body: %s", urlStr, body, res.StatusCode, resultBody))
	}

	return
}
