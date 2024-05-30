package server

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MrSwed/go-musthave-metrics/internal/server/config"
	"github.com/MrSwed/go-musthave-metrics/internal/server/domain"
	"github.com/MrSwed/go-musthave-metrics/internal/server/handler"
	"github.com/MrSwed/go-musthave-metrics/internal/server/repository"
	"github.com/MrSwed/go-musthave-metrics/internal/server/service"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type HandlerIPTestSuite struct {
	suite.Suite
	ctx context.Context
	app http.Handler
	srv *service.Service
	cfg *config.Config
}

func (suite *HandlerIPTestSuite) SetupSuite() {
	var (
		err    error
		logger *zap.Logger
	)
	suite.cfg = config.NewConfig()
	suite.cfg.TrustedSubnet = "10.17.0.0/16"
	suite.ctx = context.Background()

	repo := repository.NewRepository(&suite.cfg.StorageConfig, nil)

	suite.srv = service.NewService(repo, &suite.cfg.StorageConfig)
	logger, err = zap.NewDevelopment()
	if err != nil {
		suite.Fail(err.Error())
	}

	suite.app = handler.NewHandler(chi.NewRouter(), suite.srv, &suite.cfg.WEB, logger).Handler()
}

func TestHandlersIP(t *testing.T) {
	suite.Run(t, new(HandlerIPTestSuite))
}

func (suite *HandlerIPTestSuite) TestRequestWithXRealIp() {
	t := suite.T()

	ts := httptest.NewServer(suite.app)
	defer ts.Close()

	testGauge := domain.Gauge(1.0001)
	testGaugeName := fmt.Sprintf("testGauge%d", rand.Int())
	// save some values
	ctx := context.Background()
	_ = suite.srv.SetGauge(ctx, testGaugeName, testGauge)

	type want struct {
		code int
	}
	type args struct {
		method  string
		headers map[string]string
		path    string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Get with correct IP",
			args: args{
				method: http.MethodGet,
				headers: map[string]string{
					"X-Real-Ip": "10.17.0.10",
				},
				path: "/value/gauge/" + testGaugeName,
			},
			want: want{
				code: http.StatusOK,
			},
		},
		{
			name: "Get with incorrect IP",
			args: args{
				method: http.MethodGet,
				headers: map[string]string{
					"X-Real-Ip": "223.17.11.10",
				},
				path: "/value/gauge/" + testGaugeName,
			},
			want: want{
				code: http.StatusForbidden,
			},
		},
		{
			name: "Get without  IP",
			args: args{
				method: http.MethodGet,
				path:   "/value/gauge/" + testGaugeName,
			},
			want: want{
				code: http.StatusForbidden,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, err := http.NewRequest(test.args.method, ts.URL+test.args.path, nil)
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

			require.Equal(t, test.want.code, res.StatusCode)

		})
	}
}
