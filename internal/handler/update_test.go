// to test with real db set env DATABASE_DSN before run
// to test with file - set env FILE_STORAGE_PATH
package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MrSwed/go-musthave-metrics/internal/config"
	"github.com/MrSwed/go-musthave-metrics/internal/repository"
	"github.com/MrSwed/go-musthave-metrics/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

func NewTestUpdateConfig() (c *config.Config) {
	c = &config.Config{
		StorageConfig: config.StorageConfig{
			FileStoragePath: "",
			StorageRestore:  false,
		},
	}
	c.WithEnv().CleanSchemes()

	var err error
	if c.DatabaseDSN != "" {
		if db, err = sqlx.Open("postgres", c.DatabaseDSN); err != nil {
			log.Fatal(err)
		}
	}
	return
}

var (
	confUpd = NewTestUpdateConfig()
	dbUpd   *sqlx.DB
)

func TestUpdateMetric(t *testing.T) {
	repo := repository.NewRepository(&confUpd.StorageConfig, dbUpd)
	s := service.NewService(repo, &confUpd.StorageConfig)
	logger, _ := zap.NewDevelopment()
	h := NewHandler(s, logger).Handler()

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

func TestUpdateMetricJson(t *testing.T) {
	repo := repository.NewRepository(&confUpd.StorageConfig, dbUpd)
	s := service.NewService(repo, &confUpd.StorageConfig)
	logger, _ := zap.NewDevelopment()
	h := NewHandler(s, logger).Handler()

	ts := httptest.NewServer(h)
	defer ts.Close()

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
				response:    `{"id":"testCounter","type":"counter","delta":1}`,
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
				response:    `{"id":"testGauge","type":"gauge","value":1.1}`,
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
				response:    `{"id":"testGauge2","type":"gauge","value":0.0001}`,
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

			req, err := http.NewRequest(test.args.method, ts.URL+config.UpdateRoute, b)
			require.NoError(t, err)
			defer req.Context()

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
				assert.Equal(t, test.want.response, string(resBody))
				assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			}
		})
	}
}

func TestUpdateMetrics(t *testing.T) {
	repo := repository.NewRepository(&confUpd.StorageConfig, dbUpd)
	s := service.NewService(repo, &confUpd.StorageConfig)
	logger, _ := zap.NewDevelopment()
	h := NewHandler(s, logger).Handler()

	ts := httptest.NewServer(h)
	defer ts.Close()

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
				response:    `[{"id":"testCounter","type":"counter","delta":1},{"id":"testGauge","type":"gauge","value":100.0015}]`,
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

			req, err := http.NewRequest(test.args.method, ts.URL+config.UpdatesRoute, b)
			require.NoError(t, err)
			defer req.Context()

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
				assert.Equal(t, test.want.response, string(resBody))
				assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			}
		})
	}
}
