package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MrSwed/go-musthave-metrics/internal/server/config"
	"github.com/MrSwed/go-musthave-metrics/internal/server/constant"
	"github.com/MrSwed/go-musthave-metrics/internal/server/domain"
	"github.com/MrSwed/go-musthave-metrics/internal/server/errors"
	mocks "github.com/MrSwed/go-musthave-metrics/internal/server/mock/repository"
	"github.com/MrSwed/go-musthave-metrics/internal/server/service"
	"github.com/go-chi/chi/v5"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestMockGetMetric(t *testing.T) {
	conf := config.NewConfig()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockRepository(ctrl)

	s := service.NewService(repo, &conf.StorageConfig)

	logger, _ := zap.NewDevelopment()
	h := NewHandler(chi.NewRouter(), s, &conf.WEB, logger).Handler()
	ts := httptest.NewServer(h)
	defer ts.Close()

	testCounter := domain.Counter(1)
	testGauge := domain.Gauge(1.0001)

	_ = repo.EXPECT().GetCounter(gomock.Any(), "testCounter").Return(testCounter, nil)
	_ = repo.EXPECT().GetGauge(gomock.Any(), "testGauge").Return(testGauge, nil)
	_ = repo.EXPECT().GetGauge(gomock.Any(), gomock.Any()).Return(domain.Gauge(0), errors.ErrNotExist)
	_ = repo.EXPECT().GetCounter(gomock.Any(), gomock.Any()).Return(domain.Counter(0), errors.ErrNotExist)

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

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err, "request error")
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

func TestMockGetListMetrics(t *testing.T) {
	conf := config.NewConfig()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockRepository(ctrl)

	s := service.NewService(repo, &conf.StorageConfig)

	logger, _ := zap.NewDevelopment()
	h := NewHandler(chi.NewRouter(), s, &conf.WEB, logger).Handler()

	ts := httptest.NewServer(h)
	defer ts.Close()

	testCounter := domain.Counter(1)
	testGauge := domain.Gauge(1.0001)

	_ = repo.EXPECT().GetAllCounters(gomock.Any()).Return(domain.Counters{"testCounter": testCounter}, nil)
	_ = repo.EXPECT().GetAllGauges(gomock.Any()).Return(domain.Gauges{"testGauge": testGauge}, nil)

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

func TestMockGetMetricJson(t *testing.T) {
	conf := config.NewConfig()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockRepository(ctrl)

	s := service.NewService(repo, &conf.StorageConfig)

	logger, _ := zap.NewDevelopment()
	h := NewHandler(chi.NewRouter(), s, &conf.WEB, logger).Handler()
	ts := httptest.NewServer(h)
	defer ts.Close()

	testCounter := domain.Counter(1)
	testGauge := domain.Gauge(1.0001)

	_ = repo.EXPECT().GetCounter(gomock.Any(), "testCounter").Return(testCounter, nil)
	_ = repo.EXPECT().GetGauge(gomock.Any(), "testGauge").Return(testGauge, nil)
	_ = repo.EXPECT().GetGauge(gomock.Any(), gomock.Any()).Return(domain.Gauge(0), errors.ErrNotExist)
	_ = repo.EXPECT().GetCounter(gomock.Any(), gomock.Any()).Return(domain.Counter(0), errors.ErrNotExist)

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
					"id":   "testCounter",
					"type": "counter",
				},
			},
			want: want{
				code:        http.StatusOK,
				response:    domain.Metric{ID: "testCounter", MType: "counter", Delta: &[]domain.Counter{1}[0]},
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
				response:    domain.Metric{ID: "testGauge", MType: "gauge", Value: &[]domain.Gauge{1.0001}[0]},
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

func TestMockUpdateMetric(t *testing.T) {
	conf := config.NewConfig()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockRepository(ctrl)

	s := service.NewService(repo, &conf.StorageConfig)
	logger, _ := zap.NewDevelopment()
	h := NewHandler(chi.NewRouter(), s, &conf.WEB, logger).Handler()
	ts := httptest.NewServer(h)
	defer ts.Close()

	testCounter := domain.Counter(1)
	testGauge := domain.Gauge(1.0001)

	_ = repo.EXPECT().SetGauge(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	_ = repo.EXPECT().SetCounter(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	_ = repo.EXPECT().GetCounter(gomock.Any(), gomock.Any()).Return(testCounter, nil).AnyTimes()
	_ = repo.EXPECT().GetGauge(gomock.Any(), gomock.Any()).Return(testGauge, nil).AnyTimes()

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
			name: "SaveToFile counter. Ok",
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
			name: "SaveToFile gauge. Ok",
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
			name: "SaveToFile gauge 2. Ok",
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

func TestMockUpdateMetricJson(t *testing.T) {
	conf := config.NewConfig()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockRepository(ctrl)

	s := service.NewService(repo, &conf.StorageConfig)
	logger, _ := zap.NewDevelopment()
	h := NewHandler(chi.NewRouter(), s, &conf.WEB, logger).Handler()
	ts := httptest.NewServer(h)
	defer ts.Close()

	testCounter := domain.Counter(1)
	testGauge := domain.Gauge(1.0001)

	_ = repo.EXPECT().SetGauge(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	_ = repo.EXPECT().SetCounter(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	_ = repo.EXPECT().GetCounter(gomock.Any(), gomock.Any()).Return(testCounter, nil).AnyTimes()
	_ = repo.EXPECT().GetGauge(gomock.Any(), gomock.Any()).Return(testGauge, nil).AnyTimes()

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
			name: "SaveToFile counter. Ok",
			args: args{
				method: http.MethodPost,
				body: map[string]interface{}{
					"id":    "testCounter",
					"type":  "counter",
					"delta": testCounter,
				},
			},
			want: want{
				code:        http.StatusOK,
				response:    domain.Metric{ID: "testCounter", MType: "counter", Delta: &[]domain.Counter{1}[0]},
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			name: "SaveToFile gauge. Ok",
			args: args{
				method: http.MethodPost,
				body: map[string]interface{}{
					"id":    "testGauge",
					"type":  "gauge",
					"value": testGauge,
				},
			},
			want: want{
				code:        http.StatusOK,
				response:    domain.Metric{ID: "testGauge", MType: "gauge", Value: &[]domain.Gauge{testGauge}[0]},
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			name: "SaveToFile gauge 2. Ok",
			args: args{
				method: http.MethodPost,
				body: map[string]interface{}{
					"id":    "testGauge2",
					"type":  "gauge",
					"value": testGauge,
				},
			},
			want: want{
				code:        http.StatusOK,
				response:    domain.Metric{ID: "testGauge2", MType: "gauge", Value: &[]domain.Gauge{testGauge}[0]},
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

func TestMockUpdateMetrics(t *testing.T) {
	conf := config.NewConfig()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockRepository(ctrl)

	s := service.NewService(repo, &conf.StorageConfig)
	logger, _ := zap.NewDevelopment()
	h := NewHandler(chi.NewRouter(), s, &conf.WEB, logger).Handler()

	ts := httptest.NewServer(h)
	defer ts.Close()

	testCounter := domain.Counter(1)
	testGauge := domain.Gauge(100.0015)

	metrics := []domain.Metric{
		{
			ID:    "testCounter",
			MType: "counter",
			Delta: &testCounter,
		},
		{
			ID:    "testGauge",
			MType: "gauge",
			Value: &testGauge,
		},
	}

	_ = repo.EXPECT().SetMetrics(gomock.Any(), metrics).Return(metrics, nil).AnyTimes()

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
