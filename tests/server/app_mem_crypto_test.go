package server_test

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/MrSwed/go-musthave-metrics/internal/server/config"
	"github.com/MrSwed/go-musthave-metrics/internal/server/constant"
	"github.com/MrSwed/go-musthave-metrics/internal/server/domain"
	"github.com/MrSwed/go-musthave-metrics/internal/server/handler"
	"github.com/MrSwed/go-musthave-metrics/internal/server/repository"
	"github.com/MrSwed/go-musthave-metrics/internal/server/service"
	testHelpers "github.com/MrSwed/go-musthave-metrics/tests"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type HandlerMemCryptoTestSuite struct {
	suite.Suite
	ctx        context.Context
	app        http.Handler
	srv        *service.Service
	cfg        *config.Config
	publicKey  *rsa.PublicKey
	publicFile string
}

func (suite *HandlerMemCryptoTestSuite) loadCerts() {
	var (
		b    []byte
		cert *x509.Certificate
		err  error
	)
	b, err = os.ReadFile(suite.publicFile)
	if err != nil {
		suite.Fail(err.Error())
	}
	spkiBlock, _ := pem.Decode(b)
	cert, err = x509.ParseCertificate(spkiBlock.Bytes)
	if err == nil && (cert == nil || cert.PublicKey == nil) {
		err = errors.New("failed to load public key")
	}
	if err != nil {
		suite.Fail(err.Error())
	}
	suite.publicKey = cert.PublicKey.(*rsa.PublicKey)
	err = suite.cfg.LoadPrivateKey()
	if err != nil {
		suite.Fail(err.Error())
	}
}

func (suite *HandlerMemCryptoTestSuite) SetupSuite() {
	var (
		err    error
		logger *zap.Logger
	)
	suite.cfg = config.NewConfig()
	suite.ctx = context.Background()
	suite.cfg.CryptoKey = "/tmp/testPrivate.key"
	suite.publicFile = "/tmp/testPublic.key"
	testHelpers.CreateCertificates(suite.cfg.CryptoKey, suite.publicFile)
	suite.loadCerts()

	repo := repository.NewRepository(&suite.cfg.StorageConfig, nil)
	suite.srv = service.NewService(repo, &suite.cfg.StorageConfig)
	logger, err = zap.NewDevelopment()
	if err != nil {
		suite.Fail(err.Error())
	}

	suite.app = handler.NewHandler(suite.srv, suite.cfg, logger).HTTPHandler()
}
func (suite *HandlerMemCryptoTestSuite) TearDownSuite() {
	require.NoError(suite.T(), os.RemoveAll(suite.T().TempDir()))
}

func (suite *HandlerMemCryptoTestSuite) App() http.Handler {
	return suite.app
}
func (suite *HandlerMemCryptoTestSuite) Srv() *service.Service {
	return suite.srv
}
func (suite *HandlerMemCryptoTestSuite) DBx() *sqlx.DB {
	return nil
}
func (suite *HandlerMemCryptoTestSuite) Cfg() *config.Config {
	return suite.cfg
}
func (suite *HandlerMemCryptoTestSuite) PublicKey() *rsa.PublicKey {
	return suite.publicKey
}

func TestHandlersMemCrypto(t *testing.T) {
	suite.Run(t, new(HandlerMemCryptoTestSuite))
}

func (suite *HandlerMemCryptoTestSuite) TestUpdateMetricJson() {
	testUpdateMetricJSON(suite)
}

func (suite *HandlerMemCryptoTestSuite) TestUpdateMetrics() {
	testUpdateMetrics(suite)
}

func (suite *HandlerMemCryptoTestSuite) TestGzip() {
	testGzip(suite)
}

func (suite *HandlerMemCryptoTestSuite) TestHashKey() {
	testHashKey(suite)
}

func (suite *HandlerMemCryptoTestSuite) TestUpdateMetricsNoCrypt() {

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
		path   string
		body   interface{}
		method string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "No crypt updateMetricJSON",
			args: args{
				path:   constant.UpdatesRoute,
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
				code: http.StatusBadRequest,
			},
		},
		{
			name: "No crypt updateMetric",
			args: args{
				path:   constant.UpdateRoute,
				method: http.MethodPost,
				body: map[string]interface{}{
					"id":    testCounterName,
					"type":  "counter",
					"delta": 1,
				},
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "No crypt testGetMetricJSON",
			args: args{
				path:   constant.ValueRoute,
				method: http.MethodPost,
				body: map[string]interface{}{
					"id":   testCounterName,
					"type": "counter",
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

			req, err := http.NewRequest(test.args.method, ts.URL+test.args.path, b)
			require.NoError(t, err)

			res, err := http.DefaultClient.Do(req)
			var resBody []byte

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
