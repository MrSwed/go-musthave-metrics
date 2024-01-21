package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MrSwed/go-musthave-metrics/internal/config"
	"github.com/MrSwed/go-musthave-metrics/internal/repository"
	"github.com/MrSwed/go-musthave-metrics/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.uber.org/zap"
)

func TestGetMetric(t *testing.T) {
	conf := config.NewConfig()
	repo := repository.NewRepository(&conf.StorageConfig)
	s := service.NewService(repo)
	logger, _ := zap.NewDevelopment()
	h := NewHandler(s, &conf.ServerConfig, logger).InitRoutes()
	ts := httptest.NewServer(h.r)
	defer ts.Close()

	testCounter := int64(1)
	testGauge := 1.0001

	// save some values
	_ = s.SetGauge("testGauge", testGauge)
	_ = s.IncreaseCounter("testCounter", testCounter)

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
				path:   "/value/counter/testCounter",
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
				path:   "/value/gauge/testGauge",
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
				path:   "/value/gauge/testGauge",
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
			defer req.Context()

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
	conf := config.NewConfig()
	repo := repository.NewRepository(&conf.StorageConfig)
	s := service.NewService(repo)
	logger, _ := zap.NewDevelopment()
	h := NewHandler(s, &conf.ServerConfig, logger).InitRoutes()

	ts := httptest.NewServer(h.r)
	defer ts.Close()

	testCounter := int64(1)
	testGauge := 1.0001
	// save some values
	_ = s.SetGauge("testGauge", testGauge)
	_ = s.IncreaseCounter("testCounter", testCounter)

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
			name: "No main page",
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

			defer req.Context()

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
	conf := config.NewConfig()
	repo := repository.NewRepository(&conf.StorageConfig)
	s := service.NewService(repo)
	logger, _ := zap.NewDevelopment()
	h := NewHandler(s, &conf.ServerConfig, logger).InitRoutes()
	ts := httptest.NewServer(h.r)
	defer ts.Close()

	testCounter := int64(1)
	testGauge := 1.0001

	// save some values
	_ = s.SetGauge("testGauge", testGauge)
	_ = s.IncreaseCounter("testCounter", testCounter)

	type want struct {
		code        int
		response    string
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
					"id":   "testCounter",
					"type": "counter",
				},
			},
			want: want{
				code:        http.StatusOK,
				response:    `{"id":"testCounter","type":"counter","delta":1}`,
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			name: "Get gauge. Ok",
			args: args{
				method: http.MethodPost,
				body: map[string]interface{}{
					"id":   "testGauge",
					"type": "gauge",
				},
			},
			want: want{
				code:        http.StatusOK,
				response:    `{"id":"testGauge","type":"gauge","value":1.0001}`,
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			name: "Bad method GET",
			args: args{
				method: http.MethodGet,
				body: map[string]interface{}{
					"id":   "testGauge",
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
					"id":   "testGauge",
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
					"id": "testCounter",
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

			req, err := http.NewRequest(test.args.method, ts.URL+config.ValueRoute, b)
			require.NoError(t, err)
			defer req.Context()

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
