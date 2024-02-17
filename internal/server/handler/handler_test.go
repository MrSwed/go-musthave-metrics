// to test with real db set env DATABASE_DSN before run with created, but empty tables
// to test with file - set env FILE_STORAGE_PATH
package handler

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MrSwed/go-musthave-metrics/internal/server/config"
	"github.com/MrSwed/go-musthave-metrics/internal/server/constant"
	"github.com/MrSwed/go-musthave-metrics/internal/server/domain"
	"github.com/MrSwed/go-musthave-metrics/internal/server/repository"
	"github.com/MrSwed/go-musthave-metrics/internal/server/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

func NewConfigGetTest() (c *config.Config) {
	c = &config.Config{
		StorageConfig: config.StorageConfig{
			FileStoragePath: "",
			StorageRestore:  false,
		},
	}
	c.WithEnv().CleanSchemes()

	var err error
	if c.DatabaseDSN != "" {
		if dbTest, err = sqlx.Open("postgres", c.DatabaseDSN); err != nil {
			log.Fatal(err)
		}
	}
	return
}

var (
	confTest = NewConfigGetTest()
	dbTest   *sqlx.DB
)

func TestGetMetric(t *testing.T) {
	repo := repository.NewRepository(&confTest.StorageConfig, dbTest)
	s := service.NewService(repo, &confTest.StorageConfig)
	logger, _ := zap.NewDevelopment()
	h := NewHandler(s, &confTest.WEB, logger).Handler()
	ts := httptest.NewServer(h)
	defer ts.Close()

	testCounter := domain.Counter(1)
	testGauge := domain.Gauge(1.0001)
	testGaugeName := fmt.Sprintf("testGauge%d", rand.Int())
	testCounterName := fmt.Sprintf("testCounter%d", rand.Int())
	// save some values
	ctx := context.Background()
	_ = s.SetGauge(ctx, testGaugeName, testGauge)
	_ = s.IncreaseCounter(ctx, testCounterName, testCounter)

	type want struct {
		code        int
		response    string
		contentType string
	}
	type args struct {
		method string
		path   string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Get counter. Ok",
			args: args{
				method: http.MethodGet,
				path:   "/value/counter/" + testCounterName,
			},
			want: want{
				code:        http.StatusOK,
				response:    fmt.Sprint(testCounter),
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "Get gauge. Ok",
			args: args{
				method: http.MethodGet,
				path:   "/value/gauge/" + testGaugeName,
			},
			want: want{
				code:        http.StatusOK,
				response:    fmt.Sprint(testGauge),
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "Bad method",
			args: args{
				method: http.MethodPost,
				path:   "/value/gauge/" + testGaugeName,
			},
			want: want{
				code: http.StatusMethodNotAllowed,
			},
		},
		{
			name: "Not found Counter",
			args: args{
				method: http.MethodGet,
				path:   "/value/counter/unknownCounter",
			},
			want: want{
				code: http.StatusNotFound,
			},
		},
		{
			name: "Not found Gauge",
			args: args{
				method: http.MethodGet,
				path:   "/value/gauge/unknownGauge",
			},
			want: want{
				code: http.StatusNotFound,
			},
		},
		{
			name: "Not found - wrong path",
			args: args{
				method: http.MethodGet,
				path:   "/value/counter",
			},
			want: want{
				code: http.StatusNotFound,
			},
		},
		{
			name: "Unknown metric type",
			args: args{
				method: http.MethodGet,
				path:   "/value/unknown/testCounter",
			},
			want: want{
				code: http.StatusNotFound,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			req, err := http.NewRequest(test.args.method, ts.URL+test.args.path, nil)
			require.NoError(t, err)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			var resBody []byte

			// проверяем код ответа
			require.Equal(t, test.want.code, res.StatusCode)
			func() {
				defer func(Body io.ReadCloser) {
					err := Body.Close()
					require.NoError(t, err)
				}(res.Body)
				resBody, err = io.ReadAll(res.Body)
				require.NoError(t, err)
			}()

			if test.want.code == http.StatusOK {
				assert.Equal(t, test.want.response, string(resBody))
				assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			}
		})
	}
}

func TestGetListMetrics(t *testing.T) {
	repo := repository.NewRepository(&confTest.StorageConfig, dbTest)

	s := service.NewService(repo, &confTest.StorageConfig)
	logger, _ := zap.NewDevelopment()
	h := NewHandler(s, &confTest.WEB, logger).Handler()

	ts := httptest.NewServer(h)
	defer ts.Close()

	testCounter := domain.Counter(1)
	testGauge := domain.Gauge(1.0001)
	// save some values
	ctx := context.Background()
	_ = s.SetGauge(ctx, "testGauge", testGauge)
	_ = s.IncreaseCounter(ctx, "testCounter", testCounter)

	type want struct {
		code            int
		responseContain string
		contentType     string
	}
	type args struct {
		method string
		path   string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Get main page",
			args: args{
				method: http.MethodGet,
				path:   "/",
			},
			want: want{
				code:            http.StatusOK,
				responseContain: "testCounter",
				contentType:     "text/html; charset=utf-8",
			},
		},
		{
			name: "Not main page",
			args: args{
				method: http.MethodGet,
				path:   "/somepage",
			},
			want: want{
				code: http.StatusNotFound,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			req, err := http.NewRequest(test.args.method, ts.URL+test.args.path, nil)
			require.NoError(t, err)

			res, err := http.DefaultClient.Do(req)

			require.NoError(t, err)
			var resBody []byte

			// проверяем код ответа
			require.Equal(t, test.want.code, res.StatusCode)
			func() {
				defer func(Body io.ReadCloser) {
					err := Body.Close()
					require.NoError(t, err)
				}(res.Body)
				resBody, err = io.ReadAll(res.Body)
				require.NoError(t, err)
			}()

			if test.want.code == http.StatusOK {
				assert.Contains(t, string(resBody), "<!doctype html>")
				assert.Contains(t, string(resBody), test.want.responseContain)
				assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			}
		})
	}
}

func TestGetMetricJson(t *testing.T) {
	repo := repository.NewRepository(&confTest.StorageConfig, dbTest)
	s := service.NewService(repo, &confTest.StorageConfig)
	logger, _ := zap.NewDevelopment()
	h := NewHandler(s, &confTest.WEB, logger).Handler()
	ts := httptest.NewServer(h)
	defer ts.Close()

	testCounter := domain.Counter(1)
	testGauge := domain.Gauge(1.0001)
	testGaugeName := fmt.Sprintf("testGauge%d", rand.Int())
	testCounterName := fmt.Sprintf("testCounter%d", rand.Int())
	// save some values
	ctx := context.Background()
	_ = s.SetGauge(ctx, testGaugeName, testGauge)
	_ = s.IncreaseCounter(ctx, testCounterName, testCounter)

	type want struct {
		code        int
		response    domain.Metric
		contentType string
	}
	type args struct {
		method string
		body   interface{}
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Get counter. Ok",
			args: args{
				method: http.MethodPost,
				body: map[string]interface{}{
					"id":   testCounterName,
					"type": "counter",
				},
			},
			want: want{
				code:        http.StatusOK,
				response:    domain.Metric{ID: testCounterName, MType: "counter", Delta: &[]domain.Counter{1}[0]},
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			name: "Get gauge. Ok",
			args: args{
				method: http.MethodPost,
				body: map[string]interface{}{
					"id":   testGaugeName,
					"type": "gauge",
				},
			},
			want: want{
				code:        http.StatusOK,
				response:    domain.Metric{ID: testGaugeName, MType: "gauge", Value: &[]domain.Gauge{1.0001}[0]},
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			name: "Bad method GET",
			args: args{
				method: http.MethodGet,
				body: map[string]interface{}{
					"id":   testGaugeName,
					"type": "gauge",
				},
			},
			want: want{
				code: http.StatusMethodNotAllowed,
			},
		},
		{
			name: "Bad method PUT",
			args: args{
				method: http.MethodPut,
				body: map[string]interface{}{
					"id":   testGaugeName,
					"type": "gauge",
				},
			},
			want: want{
				code: http.StatusMethodNotAllowed,
			},
		},
		{
			name: "Not found UnknownCounter",
			args: args{
				method: http.MethodPost,
				body: map[string]interface{}{
					"id":   "UnknownCounter",
					"type": "counter",
				},
			},
			want: want{
				code: http.StatusNotFound,
			},
		},
		{
			name: "Not found UnknownGauge",
			args: args{
				method: http.MethodPost,
				body: map[string]interface{}{
					"id":   "UnknownGauge",
					"type": "gauge",
				},
			},
			want: want{
				code: http.StatusNotFound,
			},
		},
		{
			name: "Unknown metric type",
			args: args{
				method: http.MethodPost,
				body: map[string]interface{}{
					"id":   "metricName",
					"type": "unknown",
				},
			},
			want: want{
				code: http.StatusNotFound,
			},
		},
		{
			name: "Bad metric id",
			args: args{
				method: http.MethodPost,
				body: map[string]interface{}{
					"type": "counter",
				},
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Bad metric type",
			args: args{
				method: http.MethodPost,
				body: map[string]interface{}{
					"id": testCounterName,
				},
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			b := new(bytes.Buffer)
			err := json.NewEncoder(b).Encode(test.args.body)
			require.NoError(t, err)

			req, err := http.NewRequest(test.args.method, ts.URL+constant.ValueRoute, b)
			require.NoError(t, err)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			var resBody []byte

			// проверяем код ответа
			require.Equal(t, test.want.code, res.StatusCode)
			func() {
				defer func(Body io.ReadCloser) {
					err := Body.Close()
					require.NoError(t, err)
				}(res.Body)
				resBody, err = io.ReadAll(res.Body)
				require.NoError(t, err)
			}()

			if test.want.code == http.StatusOK {
				var data domain.Metric
				err = json.Unmarshal(resBody, &data)
				assert.NoError(t, err)
				assert.Equal(t, test.want.response, data)
				assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			}
		})
	}
}

func TestUpdateMetric(t *testing.T) {
	repo := repository.NewRepository(&confTest.StorageConfig, dbTest)
	s := service.NewService(repo, &confTest.StorageConfig)
	logger, _ := zap.NewDevelopment()
	h := NewHandler(s, &confTest.WEB, logger).Handler()

	ts := httptest.NewServer(h)
	defer ts.Close()

	type want struct {
		code        int
		response    string
		contentType string
	}
	type args struct {
		method string
		path   string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Save counter. Ok",
			args: args{
				method: http.MethodPost,
				path:   "/update/counter/testCounter/1",
			},
			want: want{
				code:        http.StatusOK,
				response:    "Saved: Ok",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "Save gauge. Ok",
			args: args{
				method: http.MethodPost,
				path:   "/update/gauge/testGauge/1.1",
			},
			want: want{
				code:        http.StatusOK,
				response:    "Saved: Ok",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "Save gauge 2. Ok",
			args: args{
				method: http.MethodPost,
				path:   "/update/gauge/testGauge/0.0001",
			},
			want: want{
				code:        http.StatusOK,
				response:    "Saved: Ok",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "Bad method",
			args: args{
				method: http.MethodGet,
				path:   "/update/gauge/testGauge/1.1",
			},
			want: want{
				code: http.StatusMethodNotAllowed,
			},
		},
		{
			name: "Not found 1",
			args: args{
				method: http.MethodPost,
				path:   "/update/counter/testCounter",
			},
			want: want{
				code: http.StatusNotFound,
			},
		},
		{
			name: "Not found 2",
			args: args{
				method: http.MethodPost,
				path:   "/update/counter",
			},
			want: want{
				code: http.StatusNotFound,
			},
		},
		{
			name: "Not found 3",
			args: args{
				method: http.MethodPost,
				path:   "/update/gauge",
			},
			want: want{
				code: http.StatusNotFound,
			},
		},
		{
			name: "Bad counter 1 (string)",
			args: args{
				method: http.MethodPost,
				path:   "/update/counter/testCounter/ccc",
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Bad counter 2 (float)",
			args: args{
				method: http.MethodPost,
				path:   "/update/counter/testCounter/1.1",
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Bad counter 3 (separated by a dot)",
			args: args{
				method: http.MethodPost,
				path:   "/update/counter/testCounter/1.1.1",
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Unknown metric type",
			args: args{
				method: http.MethodPost,
				path:   "/update/unknown/testCounter/122",
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Bad gauge 1 (separated by a dot)",
			args: args{
				method: http.MethodPost,
				path:   "/update/gauge/testGauge/1.1.1",
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Bad gauge 2 (string)",
			args: args{
				method: http.MethodPost,
				path:   "/update/gauge/testGauge/ddd",
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			req, err := http.NewRequest(test.args.method, ts.URL+test.args.path, nil)
			require.NoError(t, err)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			var resBody []byte
			// проверяем код ответа
			require.Equal(t, test.want.code, res.StatusCode)
			func() {
				defer func(Body io.ReadCloser) {
					err := Body.Close()
					require.NoError(t, err)
				}(res.Body)
				resBody, err = io.ReadAll(res.Body)
				require.NoError(t, err)
			}()

			if test.want.code == http.StatusOK {
				assert.Equal(t, test.want.response, string(resBody))
				assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			}
		})
	}
}

func TestUpdateMetricJson(t *testing.T) {
	repo := repository.NewRepository(&confTest.StorageConfig, dbTest)
	s := service.NewService(repo, &confTest.StorageConfig)
	logger, _ := zap.NewDevelopment()
	h := NewHandler(s, &confTest.WEB, logger).Handler()

	ts := httptest.NewServer(h)
	defer ts.Close()

	type want struct {
		code        int
		response    domain.Metric
		contentType string
	}

	type args struct {
		method string
		body   interface{}
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Save counter. Ok",
			args: args{
				method: http.MethodPost,
				body: map[string]interface{}{
					"id":    "testCounter",
					"type":  "counter",
					"delta": 1,
				},
			},
			want: want{
				code:        http.StatusOK,
				response:    domain.Metric{ID: "testCounter", MType: "counter", Delta: &[]domain.Counter{1}[0]},
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			name: "Save gauge. Ok",
			args: args{
				method: http.MethodPost,
				body: map[string]interface{}{
					"id":    "testGauge",
					"type":  "gauge",
					"value": 1.1,
				},
			},
			want: want{
				code:        http.StatusOK,
				response:    domain.Metric{ID: "testGauge", MType: "gauge", Value: &[]domain.Gauge{1.1}[0]},
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			name: "Save gauge 2. Ok",
			args: args{
				method: http.MethodPost,
				body: map[string]interface{}{
					"id":    "testGauge2",
					"type":  "gauge",
					"value": 0.0001,
				},
			},
			want: want{
				code:        http.StatusOK,
				response:    domain.Metric{ID: "testGauge2", MType: "gauge", Value: &[]domain.Gauge{0.0001}[0]},
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			name: "Bad method Get",
			args: args{
				method: http.MethodGet,
				body: map[string]interface{}{
					"id":    "testGauge",
					"type":  "gauge",
					"value": 1.1,
				},
			},
			want: want{
				code: http.StatusMethodNotAllowed,
			},
		},
		{
			name: "Bad method Put",
			args: args{
				method: http.MethodPut,
				body: map[string]interface{}{
					"id":    "testGauge",
					"type":  "gauge",
					"value": 1.1,
				},
			},
			want: want{
				code: http.StatusMethodNotAllowed,
			},
		},
		{
			name: "Bad counter 1 (string)",
			args: args{
				method: http.MethodPost,
				body: map[string]interface{}{
					"id":    "testCounter",
					"type":  "counter",
					"delta": "ccc",
				},
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "No type",
			args: args{
				method: http.MethodPost,
				body: map[string]interface{}{
					"id":    "testCounter",
					"delta": 100,
				},
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Bad counter 2 (float)",
			args: args{
				method: http.MethodPost,
				body: map[string]interface{}{
					"id":    "testCounter",
					"type":  "counter",
					"delta": 1.1,
				},
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Unknown metric type",
			args: args{
				method: http.MethodPost,
				body: map[string]interface{}{
					"id":    "testCounter",
					"type":  "unknown",
					"delta": 122,
				},
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Bad gauge 2 (string)",
			args: args{
				method: http.MethodPost,
				body: map[string]interface{}{
					"id":    "testGauge",
					"type":  "gauge",
					"delta": "122ddd",
				},
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Wrong value for gauge",
			args: args{
				method: http.MethodPost,
				body: map[string]interface{}{
					"id":    "testGauge",
					"type":  "gauge",
					"delta": 122,
				},
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Wrong value for counter",
			args: args{
				method: http.MethodPost,
				body: map[string]interface{}{
					"id":    "testGauge",
					"type":  "counter",
					"value": 122,
				},
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			b := new(bytes.Buffer)
			err := json.NewEncoder(b).Encode(test.args.body)
			require.NoError(t, err)

			req, err := http.NewRequest(test.args.method, ts.URL+constant.UpdateRoute, b)
			require.NoError(t, err)

			res, err := http.DefaultClient.Do(req)
			var resBody []byte

			// проверяем код ответа
			require.Equal(t, test.want.code, res.StatusCode)
			func() {
				defer func(Body io.ReadCloser) {
					err := Body.Close()
					require.NoError(t, err)
				}(res.Body)
				resBody, err = io.ReadAll(res.Body)
				require.NoError(t, err)
			}()

			if test.want.code == http.StatusOK {
				var data domain.Metric
				err = json.Unmarshal(resBody, &data)
				assert.NoError(t, err)
				assert.Equal(t, test.want.response, data)
				assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			}
		})
	}
}

func TestUpdateMetrics(t *testing.T) {
	repo := repository.NewRepository(&confTest.StorageConfig, dbTest)
	s := service.NewService(repo, &confTest.StorageConfig)
	logger, _ := zap.NewDevelopment()
	h := NewHandler(s, &confTest.WEB, logger).Handler()

	ts := httptest.NewServer(h)
	defer ts.Close()

	type want struct {
		code        int
		response    []domain.Metric
		contentType string
	}

	type args struct {
		method string
		body   interface{}
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Save. Ok",
			args: args{
				method: http.MethodPost,
				body: []map[string]interface{}{
					{
						"id":    "testCounter",
						"type":  "counter",
						"delta": 1,
					},
					{
						"id":    "testGauge",
						"type":  "gauge",
						"value": 100.0015,
					},
				},
			},
			want: want{
				code:        http.StatusOK,
				response:    []domain.Metric{{ID: "testCounter", MType: "counter", Delta: &[]domain.Counter{1}[0]}, {ID: "testGauge", MType: "gauge", Value: &[]domain.Gauge{100.0015}[0]}},
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			name: "Bad metric type",
			args: args{
				method: http.MethodPost,
				body: []map[string]interface{}{
					{
						"id":    "testCounter",
						"type":  "unknownType",
						"delta": 1,
					},
					{
						"id":    "testGauge",
						"type":  "gauge",
						"value": 100.0015,
					},
				},
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "No type",
			args: args{
				method: http.MethodPost,
				body: []map[string]interface{}{
					{
						"id":    "testCounter",
						"delta": 1,
					},
					{
						"id":    "testGauge",
						"type":  "gauge",
						"value": 100.0015,
					},
				},
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Bad counter 2 (float)",
			args: args{
				method: http.MethodPost,
				body: []map[string]interface{}{
					{
						"id":    "testCounter",
						"type":  "unknownType",
						"delta": 1.1,
					},
					{
						"id":    "testGauge",
						"type":  "gauge",
						"value": 100.0015,
					},
				},
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Bad value",
			args: args{
				method: http.MethodPost,
				body: []map[string]interface{}{
					{
						"id":    "testCounter",
						"type":  "unknownType",
						"delta": "ddd",
					},
					{
						"id":    "testGauge",
						"type":  "gauge",
						"value": "ddd",
					},
				},
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Wrong types",
			args: args{
				method: http.MethodPost,
				body: []map[string]interface{}{
					{
						"id":    "testCounter",
						"type":  "gauge",
						"delta": 1,
					},
					{
						"id":    "testGauge",
						"type":  "counter",
						"value": 100.0015,
					},
				},
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			b := new(bytes.Buffer)
			err := json.NewEncoder(b).Encode(test.args.body)
			require.NoError(t, err)

			req, err := http.NewRequest(test.args.method, ts.URL+constant.UpdatesRoute, b)
			require.NoError(t, err)

			res, err := http.DefaultClient.Do(req)
			var resBody []byte

			// проверяем код ответа
			require.Equal(t, test.want.code, res.StatusCode)
			func() {
				defer func(Body io.ReadCloser) {
					err := Body.Close()
					require.NoError(t, err)
				}(res.Body)
				resBody, err = io.ReadAll(res.Body)
				require.NoError(t, err)
			}()

			if test.want.code == http.StatusOK {
				var data []domain.Metric
				err = json.Unmarshal(resBody, &data)
				assert.NoError(t, err)
				assert.Equal(t, test.want.response, data)
				assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			}
		})
	}
}

func TestGzip(t *testing.T) {
	repo := repository.NewRepository(&confTest.StorageConfig, dbTest)
	s := service.NewService(repo, &confTest.StorageConfig)
	logger, _ := zap.NewDevelopment()
	h := NewHandler(s, &confTest.WEB, logger).Handler()

	ts := httptest.NewServer(h)
	defer ts.Close()

	type want struct {
		code        int
		response    []domain.Metric
		contentType string
		headers     map[string]string
	}
	type args struct {
		method  string
		headers map[string]string
		body    interface{}
	}
	tests := []struct {
		name string
		args args
		want want
	}{

		{
			name: "Gzip compress answer at save json. Ok",
			args: args{
				method: http.MethodPost,
				body: []map[string]interface{}{
					{
						"id":    "testCounter1",
						"type":  "counter",
						"delta": 1,
					},
					{
						"id":    "testGauge",
						"type":  "gauge",
						"value": 100.0015,
					},
				},
				headers: map[string]string{
					"Accept-Encoding": "gzip",
				},
			},
			want: want{
				code:        http.StatusOK,
				response:    []domain.Metric{{ID: "testCounter1", MType: "counter", Delta: &[]domain.Counter{1}[0]}, {ID: "testGauge", MType: "gauge", Value: &[]domain.Gauge{100.0015}[0]}},
				contentType: "application/json; charset=utf-8",
				headers: map[string]string{
					"Content-Encoding": "gzip",
				},
			},
		},
		{
			name: "Gzip decompress request save json. Ok",
			args: args{
				method: http.MethodPost,
				body: []map[string]interface{}{
					{
						"id":    "testCounter2",
						"type":  "counter",
						"delta": 1,
					},
					{
						"id":    "testGauge",
						"type":  "gauge",
						"value": 100.0015,
					},
				},
				headers: map[string]string{
					"Content-Encoding": "gzip",
				},
			},
			want: want{
				code:        http.StatusOK,
				response:    []domain.Metric{{ID: "testCounter2", MType: "counter", Delta: &[]domain.Counter{1}[0]}, {ID: "testGauge", MType: "gauge", Value: &[]domain.Gauge{100.0015}[0]}},
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			name: "Gzip compress/decompress answer/request save json. Ok",
			args: args{
				method: http.MethodPost,
				body: []map[string]interface{}{
					{
						"id":    "testCounter3",
						"type":  "counter",
						"delta": 1,
					},
					{
						"id":    "testGauge",
						"type":  "gauge",
						"value": 100.0015,
					},
				},
				headers: map[string]string{
					"Accept-Encoding":  "gzip",
					"Content-Encoding": "gzip",
				},
			},
			want: want{
				code:        http.StatusOK,
				response:    []domain.Metric{{ID: "testCounter3", MType: "counter", Delta: &[]domain.Counter{1}[0]}, {ID: "testGauge", MType: "gauge", Value: &[]domain.Gauge{100.0015}[0]}},
				contentType: "application/json; charset=utf-8",
				headers: map[string]string{
					"Content-Encoding": "gzip",
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			b := new(bytes.Buffer)
			err := json.NewEncoder(b).Encode(test.args.body)
			require.NoError(t, err)
			if len(test.args.headers) > 0 && test.args.headers["Content-Encoding"] == "gzip" {
				compB := new(bytes.Buffer)
				w := gzip.NewWriter(compB)
				_, err = w.Write(b.Bytes())
				b = compB
				require.NoError(t, err)

				err = w.Flush()
				require.NoError(t, err)

				err = w.Close()
				require.NoError(t, err)
			}

			req, err := http.NewRequest(test.args.method, ts.URL+constant.UpdatesRoute, b)
			require.NoError(t, err)

			for k, v := range test.args.headers {
				req.Header.Add(k, v)
			}

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer func() {
				err := res.Body.Close()
				require.NoError(t, err)
			}()

			var resBody []byte

			// проверяем код ответа
			require.Equal(t, test.want.code, res.StatusCode)
			resBody, err = io.ReadAll(res.Body)
			require.NoError(t, err)

			for k, v := range test.want.headers {
				assert.True(t, res.Header.Get(k) == v)
			}

			if len(test.want.response) > 0 {
				assert.True(t, len(resBody) > 0)
				if len(test.args.headers) > 0 && test.args.headers["Accept-Encoding"] == "gzip" {
					b := bytes.NewBuffer(resBody)
					r, err := gzip.NewReader(b)
					if !errors.Is(err, io.EOF) {
						require.NoError(t, err)
					}
					var resB bytes.Buffer
					_, err = resB.ReadFrom(r)
					require.NoError(t, err)

					resBody = resB.Bytes()
					err = r.Close()
					require.NoError(t, err)
				}
				var data []domain.Metric
				err = json.Unmarshal(resBody, &data)
				assert.NoError(t, err)
				assert.Equal(t, test.want.response, data)
			}

			if test.want.contentType != "" {
				assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			}
		})
	}
}
