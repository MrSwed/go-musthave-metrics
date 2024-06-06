package server_test

import (
	"bytes"
	"compress/gzip"
	"context"
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"go-musthave-metrics/internal/server/handler/rest"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-musthave-metrics/internal/server/config"
	"go-musthave-metrics/internal/server/constant"
	"go-musthave-metrics/internal/server/domain"
	errM "go-musthave-metrics/internal/server/migrate"
	"go-musthave-metrics/internal/server/service"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type HandlerTestSuite interface {
	App() http.Handler
	Srv() *service.Service
	T() *testing.T
	DBx() *sqlx.DB
	Cfg() *config.Config
	PublicKey() *rsa.PublicKey
}

func testMigrate(suite HandlerTestSuite) {
	t := suite.T()
	t.Run("Migrate", func(t *testing.T) {
		_, err := errM.Migrate(suite.DBx().DB)
		switch {
		case errors.Is(err, migrate.ErrNoChange):
		default:
			require.NoError(t, err)
		}
	})
}

func testGetMetric(suite HandlerTestSuite) {
	t := suite.T()

	ts := httptest.NewServer(suite.App())
	defer ts.Close()

	testCounter := domain.Counter(1)
	testGauge := domain.Gauge(1.0001)
	testGaugeName := fmt.Sprintf("testGauge%d", rand.Int())
	testCounterName := fmt.Sprintf("testCounter%d", rand.Int())
	// save some values
	ctx := context.Background()
	_ = suite.Srv().SetGauge(ctx, testGaugeName, testGauge)
	_ = suite.Srv().IncreaseCounter(ctx, testCounterName, testCounter)

	type want struct {
		response    string
		contentType string
		code        int
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
					err = Body.Close()
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

func testGetListMetrics(suite HandlerTestSuite) {
	t := suite.T()

	ts := httptest.NewServer(suite.App())
	defer ts.Close()

	testCounter := domain.Counter(1)
	testGauge := domain.Gauge(1.0001)
	// save some values
	ctx := context.Background()
	_ = suite.Srv().SetGauge(ctx, "testGauge", testGauge)
	_ = suite.Srv().IncreaseCounter(ctx, "testCounter", testCounter)

	type want struct {
		responseContain string
		contentType     string
		code            int
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
					err = Body.Close()
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

func testGetMetricJSON(suite HandlerTestSuite) {
	t := suite.T()

	ts := httptest.NewServer(suite.App())
	defer ts.Close()

	testCounter := domain.Counter(1)
	testGauge := domain.Gauge(1.0001)
	testGaugeName := fmt.Sprintf("testGauge%d", rand.Int())
	testCounterName := fmt.Sprintf("testCounter%d", rand.Int())
	// save some values
	ctx := context.Background()
	_ = suite.Srv().SetGauge(ctx, testGaugeName, testGauge)
	_ = suite.Srv().IncreaseCounter(ctx, testCounterName, testCounter)

	type want struct {
		response    domain.Metric
		contentType string
		code        int
	}
	type args struct {
		body   interface{}
		method string
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

			maybeCryptBody(b, suite.PublicKey())
			req, err := http.NewRequest(test.args.method, ts.URL+constant.ValueRoute, b)
			require.NoError(t, err)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			var resBody []byte

			// проверяем код ответа
			require.Equal(t, test.want.code, res.StatusCode)
			func() {
				defer func(Body io.ReadCloser) {
					err = Body.Close()
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

func testUpdateMetric(suite HandlerTestSuite) {
	t := suite.T()

	ts := httptest.NewServer(suite.App())
	defer ts.Close()

	type want struct {
		response    string
		contentType string
		code        int
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
					err = Body.Close()
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

func testUpdateMetricJSON(suite HandlerTestSuite) {
	t := suite.T()

	ts := httptest.NewServer(suite.App())
	defer ts.Close()

	testCounterName := fmt.Sprintf("testCounter%d", rand.Int())
	type want struct {
		response    domain.Metric
		contentType string
		code        int
	}
	type args struct {
		body   interface{}
		method string
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
					"id":    testCounterName,
					"type":  "counter",
					"delta": 1,
				},
			},
			want: want{
				code:        http.StatusOK,
				response:    domain.Metric{ID: testCounterName, MType: "counter", Delta: &[]domain.Counter{1}[0]},
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
					"id":    testCounterName,
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
					"id":    testCounterName,
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
					"id":    testCounterName,
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
					"id":    testCounterName,
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

			maybeCryptBody(b, suite.PublicKey())
			req, err := http.NewRequest(test.args.method, ts.URL+constant.UpdateRoute, b)
			require.NoError(t, err)

			res, err := http.DefaultClient.Do(req)
			var resBody []byte

			// проверяем код ответа
			require.Equal(t, test.want.code, res.StatusCode)
			func() {
				defer func(Body io.ReadCloser) {
					err = Body.Close()
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

func testUpdateMetrics(suite HandlerTestSuite) {
	t := suite.T()

	ts := httptest.NewServer(suite.App())
	defer ts.Close()

	testCounterName := fmt.Sprintf("testCounter%d", rand.Int())

	type want struct {
		contentType string
		response    []domain.Metric
		code        int
	}

	type args struct {
		body   interface{}
		method string
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
						"id":    testCounterName,
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
				response:    []domain.Metric{{ID: testCounterName, MType: "counter", Delta: &[]domain.Counter{1}[0]}, {ID: "testGauge", MType: "gauge", Value: &[]domain.Gauge{100.0015}[0]}},
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			name: "Bad metric type",
			args: args{
				method: http.MethodPost,
				body: []map[string]interface{}{
					{
						"id":    testCounterName,
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
						"id":    testCounterName,
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
						"id":    testCounterName,
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
						"id":    testCounterName,
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
						"id":    testCounterName,
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

			maybeCryptBody(b, suite.PublicKey())
			req, err := http.NewRequest(test.args.method, ts.URL+constant.UpdatesRoute, b)
			require.NoError(t, err)

			res, err := http.DefaultClient.Do(req)
			var resBody []byte

			// проверяем код ответа
			require.Equal(t, test.want.code, res.StatusCode)
			func() {
				defer func(Body io.ReadCloser) {
					err = Body.Close()
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

func testGzip(suite HandlerTestSuite) {
	t := suite.T()

	ts := httptest.NewServer(suite.App())
	defer ts.Close()

	testCounterName1 := fmt.Sprintf("testCounter%d", rand.Int())
	testCounterName2 := fmt.Sprintf("testCounter%d", rand.Int())
	testCounterName3 := fmt.Sprintf("testCounter%d", rand.Int())
	testCounterName4 := fmt.Sprintf("testCounter%d", rand.Int())

	type want struct {
		headers     map[string]string
		contentType string
		response    []domain.Metric
		code        int
	}
	type args struct {
		body    interface{}
		headers map[string]string
		method  string
		noGzip  bool
	}
	tests := []struct {
		args args
		name string
		want want
	}{
		{
			name: "Gzip compress answer at save json. Ok",
			args: args{
				method: http.MethodPost,
				body: []map[string]interface{}{
					{
						"id":    testCounterName1,
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
				response:    []domain.Metric{{ID: testCounterName1, MType: "counter", Delta: &[]domain.Counter{1}[0]}, {ID: "testGauge", MType: "gauge", Value: &[]domain.Gauge{100.0015}[0]}},
				contentType: "application/json; charset=utf-8",
				headers: map[string]string{
					"Content-Encoding": "gzip",
				},
			},
		},
		{
			name: "Gzip send header, but no compress, StatusOk",
			args: args{
				noGzip: true,
				method: http.MethodPost,
				body: []map[string]interface{}{
					{
						"id":    testCounterName4,
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
					"Accept-Encoding":  "gzip",
				},
			},
			want: want{
				code:        http.StatusOK,
				response:    []domain.Metric{{ID: testCounterName4, MType: "counter", Delta: &[]domain.Counter{1}[0]}, {ID: "testGauge", MType: "gauge", Value: &[]domain.Gauge{100.0015}[0]}},
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
						"id":    testCounterName2,
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
				response:    []domain.Metric{{ID: testCounterName2, MType: "counter", Delta: &[]domain.Counter{1}[0]}, {ID: "testGauge", MType: "gauge", Value: &[]domain.Gauge{100.0015}[0]}},
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			name: "Gzip compress/decompress answer/request save json. Ok",
			args: args{
				method: http.MethodPost,
				body: []map[string]interface{}{
					{
						"id":    testCounterName3,
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
				response:    []domain.Metric{{ID: testCounterName3, MType: "counter", Delta: &[]domain.Counter{1}[0]}, {ID: "testGauge", MType: "gauge", Value: &[]domain.Gauge{100.0015}[0]}},
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
			if len(test.args.headers) > 0 && test.args.headers["Content-Encoding"] == "gzip" && !test.args.noGzip {
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

			maybeCryptBody(b, suite.PublicKey())
			req, err := http.NewRequest(test.args.method, ts.URL+constant.UpdatesRoute, b)
			require.NoError(t, err)

			for k, v := range test.args.headers {
				req.Header.Add(k, v)
			}

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer func() {
				err = res.Body.Close()
				require.NoError(t, err)
			}()

			var resBody []byte

			// проверяем код ответа
			require.Equal(t, test.want.code, res.StatusCode)
			resBody, err = io.ReadAll(res.Body)
			require.NoError(t, err)

			for k, v := range test.want.headers {
				assert.True(t, res.Header.Get(k) == v,
					fmt.Sprintf("expected %v header: %s actual: %s", k, v, res.Header.Get(k)))
			}

			if len(test.want.response) > 0 {
				assert.True(t, len(resBody) > 0)
				if len(test.args.headers) > 0 && test.args.headers["Accept-Encoding"] == "gzip" {
					b := bytes.NewBuffer(resBody)
					var r *gzip.Reader
					r, err = gzip.NewReader(b)
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

func testHashKey(suite HandlerTestSuite) {
	t := suite.T()
	suite.Cfg().Key = "secretKey"
	ts := httptest.NewServer(suite.App())
	defer func() {
		ts.Close()
		suite.Cfg().Key = ""
	}()

	data1 := []domain.Metric{{ID: fmt.Sprintf("testCounter%d", rand.Int()), MType: "counter", Delta: &[]domain.Counter{1}[0]}, {ID: "testGauge", MType: "gauge", Value: &[]domain.Gauge{100.0015}[0]}}
	dataBody1, err := json.Marshal(data1)
	require.NoError(t, err)

	data2 := []domain.Metric{{ID: fmt.Sprintf("testCounter%d", rand.Int()), MType: "counter", Delta: &[]domain.Counter{1}[0]}, {ID: "testGauge", MType: "gauge", Value: &[]domain.Gauge{100.0015}[0]}}
	dataBody2, err := json.Marshal(data2)
	require.NoError(t, err)

	data3 := []domain.Metric{{ID: fmt.Sprintf("testCounter%d", rand.Int()), MType: "counter", Delta: &[]domain.Counter{1}[0]}, {ID: "testGauge", MType: "gauge", Value: &[]domain.Gauge{100.0015}[0]}}
	dataBody3, err := json.Marshal(data3)
	require.NoError(t, err)

	type want struct {
		headers     map[string]string
		contentType string
		response    []domain.Metric
		code        int
	}
	type args struct {
		method  string
		headers map[string]string
		body    []byte
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Save json right secret key. Ok",
			args: args{
				method: http.MethodPost,
				body:   dataBody1,
				headers: map[string]string{
					constant.HeaderSignKey: rest.SignData(suite.Cfg().Key, dataBody1),
				},
			},
			want: want{
				code:        http.StatusOK,
				response:    data1,
				contentType: "application/json; charset=utf-8",
				headers: map[string]string{
					constant.HeaderSignKey: rest.SignData(suite.Cfg().Key, dataBody1),
				},
			},
		},
		{
			name: "Save json right secret key (gzip). Ok",
			args: args{
				method: http.MethodPost,
				body:   dataBody2,
				headers: map[string]string{
					"Accept-Encoding":      "gzip",
					"Content-Encoding":     "gzip",
					constant.HeaderSignKey: rest.SignData(suite.Cfg().Key, dataBody2),
				},
			},
			want: want{
				code:        http.StatusOK,
				response:    data2,
				contentType: "application/json; charset=utf-8",
				headers: map[string]string{
					"Content-Encoding":     "gzip",
					constant.HeaderSignKey: rest.SignData(suite.Cfg().Key, dataBody2),
				},
			},
		},
		{
			name: "Save json. WRONG secret key. ",
			args: args{
				method: http.MethodPost,
				body:   dataBody1,
				headers: map[string]string{
					constant.HeaderSignKey: rest.SignData("wrong secret key", dataBody1),
				},
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Save json. BAD secret key. ",
			args: args{
				method: http.MethodPost,
				body:   dataBody1,
				headers: map[string]string{
					constant.HeaderSignKey: "bad secret data key",
				},
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Save json, WRONG secret key (gzip).",
			args: args{
				method: http.MethodPost,
				body:   dataBody2,
				headers: map[string]string{
					"Accept-Encoding":      "gzip",
					"Content-Encoding":     "gzip",
					constant.HeaderSignKey: rest.SignData("wrong secret key", dataBody2),
				},
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Save json, NO secret key (gzip).",
			args: args{
				method: http.MethodPost,
				body:   dataBody3,
				headers: map[string]string{
					"Accept-Encoding":  "gzip",
					"Content-Encoding": "gzip",
				},
			},
			want: want{
				code:        http.StatusOK,
				response:    data3,
				contentType: "application/json; charset=utf-8",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			b := new(bytes.Buffer)
			b.Write(test.args.body)
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

			maybeCryptBody(b, suite.PublicKey())
			req, err := http.NewRequest(test.args.method, ts.URL+constant.UpdatesRoute, b)
			require.NoError(t, err)

			for k, v := range test.args.headers {
				req.Header.Add(k, v)
			}
			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer func() {
				err = res.Body.Close()
				require.NoError(t, err)
			}()

			var resBody []byte

			// проверяем код ответа
			require.Equal(t, test.want.code, res.StatusCode)
			resBody, err = io.ReadAll(res.Body)
			require.NoError(t, err)

			for k, v := range test.want.headers {
				assert.True(t, res.Header.Get(k) == v, fmt.Sprintf("want %s, get %s", v, res.Header.Get(k)))
			}

			if len(test.want.response) > 0 {
				assert.True(t, len(resBody) > 0)
				if len(test.args.headers) > 0 && test.args.headers["Accept-Encoding"] == "gzip" {
					b := bytes.NewBuffer(resBody)
					var r *gzip.Reader
					r, err = gzip.NewReader(b)
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

func testPing(suite HandlerTestSuite) {
	t := suite.T()
	ts := httptest.NewServer(suite.App())
	defer func() {
		ts.Close()
	}()

	type want struct {
		headers     map[string]string
		contentType string
		response    []byte
		code        int
	}
	type args struct {
		headers map[string]string
		method  string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Ping",
			args: args{
				method: http.MethodGet,
			},
			want: want{
				code:        http.StatusOK,
				response:    []byte("Status: ok"),
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "Ping, wrong method",
			args: args{
				method: http.MethodPost,
			},
			want: want{
				code: http.StatusMethodNotAllowed,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			req, err := http.NewRequest(test.args.method, ts.URL+"/ping", nil)
			require.NoError(t, err)

			for k, v := range test.args.headers {
				req.Header.Add(k, v)
			}
			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer func() {
				err = res.Body.Close()
				require.NoError(t, err)
			}()

			var resBody []byte

			// проверяем код ответа
			require.Equal(t, test.want.code, res.StatusCode)
			resBody, err = io.ReadAll(res.Body)
			require.NoError(t, err)

			for k, v := range test.want.headers {
				assert.True(t, res.Header.Get(k) == v, fmt.Sprintf("want %s, get %s", v, res.Header.Get(k)))
			}

			if len(test.want.response) > 0 {
				assert.True(t, len(resBody) > 0)
				if len(test.args.headers) > 0 && test.args.headers["Accept-Encoding"] == "gzip" {
					b := bytes.NewBuffer(resBody)
					var r *gzip.Reader
					r, err = gzip.NewReader(b)
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
				assert.Equal(t, test.want.response, resBody)
			}

			if test.want.contentType != "" {
				assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			}
		})
	}
}

func testSaveToFile(suite HandlerTestSuite, suiteCtx context.Context) {
	t := suite.T()
	testCounter := domain.Counter(1)
	testGauge := domain.Gauge(1.0001)
	testGaugeName := fmt.Sprintf("testGauge%d", rand.Int())
	testCounterName := fmt.Sprintf("testCounter%d", rand.Int())
	// save some values
	ctx := context.Background()
	_ = suite.Srv().SetGauge(ctx, testGaugeName, testGauge)
	_ = suite.Srv().IncreaseCounter(ctx, testCounterName, testCounter)

	t.Run("Save file", func(t *testing.T) {
		_, err := suite.Srv().SaveToFile(suiteCtx)
		require.NoError(t, err)
	})
}

func maybeCryptBody(bodyBuf *bytes.Buffer, publicKey *rsa.PublicKey) {
	if publicKey != nil {
		cipherBody, err := rsa.EncryptOAEP(sha256.New(), crand.Reader, publicKey, bodyBuf.Bytes(), nil)
		bodyBuf.Reset()
		if err != nil {
			bodyBuf.WriteString(err.Error())
			return
		}
		bodyBuf.Write(cipherBody)
	}
}
