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
	mocks "github.com/MrSwed/go-musthave-metrics/internal/mock/repository"
	"github.com/MrSwed/go-musthave-metrics/internal/service"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestMockUpdateMetric(t *testing.T) {
	conf := config.NewConfig()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockRepository(ctrl)

	s := service.NewService(repo, &conf.StorageConfig)
	logger, _ := zap.NewDevelopment()
	h := NewHandler(s, logger).Handler()
	ts := httptest.NewServer(h)
	defer ts.Close()

	testCounter := int64(1)
	testGauge := 1.0001

	_ = repo.EXPECT().SetGauge(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	_ = repo.EXPECT().SetCounter(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	_ = repo.EXPECT().GetCounter(gomock.Any()).Return(testCounter, nil).AnyTimes()
	_ = repo.EXPECT().GetGauge(gomock.Any()).Return(testGauge, nil).AnyTimes()

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

func TestMockUpdateMetricJson(t *testing.T) {
	conf := config.NewConfig()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockRepository(ctrl)

	s := service.NewService(repo, &conf.StorageConfig)
	logger, _ := zap.NewDevelopment()
	h := NewHandler(s, logger).Handler()
	ts := httptest.NewServer(h)
	defer ts.Close()

	testCounter := int64(1)
	testGauge := 1.0001

	_ = repo.EXPECT().SetGauge(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	_ = repo.EXPECT().SetCounter(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	_ = repo.EXPECT().GetCounter(gomock.Any()).Return(testCounter, nil).AnyTimes()
	_ = repo.EXPECT().GetGauge(gomock.Any()).Return(testGauge, nil).AnyTimes()

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
				response:    `{"id":"testCounter","type":"counter","delta":1}`,
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
				response:    fmt.Sprintf(`{"id":"testGauge","type":"gauge","value":%0.4f}`, testGauge),
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
				response:    fmt.Sprintf(`{"id":"testGauge2","type":"gauge","value":%0.4f}`, testGauge),
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
