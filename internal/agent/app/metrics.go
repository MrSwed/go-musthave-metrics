package app

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"reflect"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/MrSwed/go-musthave-metrics/internal/agent/config"
	"github.com/MrSwed/go-musthave-metrics/internal/agent/constant"
	myErr "github.com/MrSwed/go-musthave-metrics/internal/agent/error"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"golang.org/x/sync/errgroup"
)

// MetricsCollects metrics collection
type MetricsCollects struct {
	c              *config.Config
	CPUutilization []float64
	runtime.MemStats
	PollCount   int64
	RandomValue float64
	TotalMemory float64
	FreeMemory  float64
	m           sync.RWMutex
}

func NewMetricsCollects(c *config.Config) *MetricsCollects {
	return &MetricsCollects{
		c: c,
	}
}

// GetMetrics reg runtime MemStats metrics
func (m *MetricsCollects) GetMetrics() {
	m.m.Lock()
	defer m.m.Unlock()
	runtime.ReadMemStats(&m.MemStats)
	m.PollCount++
	m.RandomValue = rand.Float64()
}

// GetGopMetrics get memory and cpu metrics by gopsutil
func (m *MetricsCollects) GetGopMetrics() error {
	m.m.Lock()
	defer m.m.Unlock()
	vmst, err := mem.VirtualMemory()
	if err != nil {
		return err
	}
	m.FreeMemory = float64(vmst.Free)
	m.TotalMemory = float64(vmst.Total)

	m.CPUutilization, err = cpu.Percent(time.Second*1, true)
	if err != nil {
		return err
	}
	return nil
}

// ListMetrics create metrics collection
func (m *MetricsCollects) ListMetrics() (metrics []*Metric, err error) {
	var er error

	mRefVal := reflect.Indirect(reflect.ValueOf(m))
	lRefVal := reflect.ValueOf(m.c.MetricLists)
	lRefType := reflect.TypeOf(m.c.MetricLists)
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
				var v interface{}
				if refV := mRefVal.FieldByName(mName); refV.IsValid() {
					m.m.RLock()
					v = refV.Interface()
					m.m.RUnlock()
				} else {
					err = errors.Join(err, myErr.ErrWrap(fmt.Errorf("unknown metric name %s", mName)))
					continue
				}
				switch g := v.(type) {
				case []float64:
					for idx := range g {
						oneMetric := NewMetric(mName+strconv.FormatInt(int64(idx+1), 10), mType)
						if er = oneMetric.Set(g[i]); er != nil {
							err = errors.Join(err, myErr.ErrWrap(er))
							continue
						}
						metrics = append(metrics, oneMetric)
					}
				default:
					oneMetric := NewMetric(mName, mType)
					if er = oneMetric.Set(v); er != nil {
						err = errors.Join(err, myErr.ErrWrap(er))
						continue
					}
					metrics = append(metrics, oneMetric)
				}
			}
		}
	}
	return
}

// SendMetrics send metrics collection
func (m *MetricsCollects) SendMetrics(ctx context.Context) (n int, err error) {
	var (
		metrics []*Metric
		er      error
		wg      sync.WaitGroup
		g       *errgroup.Group
	)

	if metrics, er = m.ListMetrics(); er != nil {
		err = errors.Join(err, myErr.ErrWrap(er))
	}
	n = len(metrics)

	semaphore := NewSemaphore(m.c.RateLimit)
	sendCount := 1
	if m.c.SendSize > 0 && m.c.SendSize < len(metrics) {
		sendCount = len(metrics) / m.c.SendSize
		if len(metrics)%m.c.SendSize > 0 {
			sendCount++
		}
	}
	g, ctx = errgroup.WithContext(ctx)
	for s := 0; s < sendCount; s++ {
		select {
		case <-ctx.Done():
			log.Print("ctx done, do not send parts")
			return
		default:
		}
		wg.Add(1)
		start, finish := s*m.c.SendSize, (s+1)*m.c.SendSize
		if finish > len(metrics) || finish == 0 {
			finish = len(metrics)
		}
		g.Go(func() (err error) {
			func(metrics []*Metric) {
				semaphore.Acquire()
				defer wg.Done()
				defer semaphore.Release()
				err = m.request(metrics)
			}(metrics[start:finish])
			return err
		})
	}
	if er = g.Wait(); er != nil {
		err = errors.Join(err, myErr.ErrWrap(er))
	}
	return
}

func (m *MetricsCollects) request(metrics []*Metric) (err error) {
	var er error

	// data to json body
	var body []byte
	if body, er = json.Marshal(metrics); er != nil {
		err = errors.Join(err, myErr.ErrWrap(er))
		return
	}

	// compress stage
	compressedBody := new(bytes.Buffer)
	zb := gzip.NewWriter(compressedBody)
	if _, er = zb.Write(body); er != nil {
		err = errors.Join(err, myErr.ErrWrap(er))
		return
	}

	urlStr := m.c.ServerAddress + constant.BaseURL
	if er = zb.Close(); er != nil {
		err = errors.Join(err, myErr.ErrWrap(er))
		return
	}

	// prepare request
	var req *http.Request
	if req, er = http.NewRequest("POST", urlStr, compressedBody); er != nil {
		err = errors.Join(err, myErr.ErrWrap(er))
		return
	}
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	// sign at header
	if m.c != nil && m.c.Key != "" {
		h := hmac.New(sha256.New, []byte(m.c.Key))
		if _, err = h.Write(body); err != nil {
			err = errors.Join(err, myErr.ErrWrap(er))
			return
		}
		req.Header.Set(constant.HeaderSignKey, hex.EncodeToString(h.Sum(nil)))
	}

	// do request
	var res *http.Response
	if res, er = http.DefaultClient.Do(req); er != nil {
		err = errors.Join(err, myErr.ErrWrap(er))
		return
	}

	// read result
	var resultBody []byte
	if resultBody, er = io.ReadAll(res.Body); er != nil {
		err = errors.Join(err, myErr.ErrWrap(er))
	}
	if er = res.Body.Close(); er != nil {
		err = errors.Join(err, myErr.ErrWrap(er))
	}
	if res.StatusCode != http.StatusOK {
		err = errors.Join(err, fmt.Errorf("post to %s with body: %s. Get: statusCode: %d;  answer body: %s", urlStr, body, res.StatusCode, resultBody))
	}

	return
}
