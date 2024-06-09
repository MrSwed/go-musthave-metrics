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
	"go-musthave-metrics/internal/server/app"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go-musthave-metrics/internal/server/config"
	"go-musthave-metrics/internal/server/constant"
	"go-musthave-metrics/internal/server/domain"
	"go-musthave-metrics/internal/server/repository"
	"go-musthave-metrics/internal/server/service"
	testHelpers "go-musthave-metrics/tests"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type HandlerMemCryptoTestSuite struct {
	suite.Suite
	ctx        context.Context
	stop       context.CancelFunc
	srv        *service.Service
	cfg        *config.Config
	a          *app.App
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
		err error
	)

	suite.cfg = config.NewConfig()
	suite.ctx, suite.stop = context.WithCancel(context.Background())
	suite.cfg.StorageConfig.FileStoragePath = filepath.Join(suite.T().TempDir(), fmt.Sprintf("store-data-%d.json", rand.Intn(200000)))
	suite.cfg.Address = net.JoinHostPort("localhost", fmt.Sprintf("%d", rand.Intn(200)+20000))
	suite.cfg.GRPCAddress = net.JoinHostPort("", fmt.Sprintf("%d", rand.Intn(200)+30000))

	suite.cfg.CryptoKey = "/tmp/testPrivate.key"
	suite.publicFile = "/tmp/testPublic.key"
	testHelpers.CreateCertificates(suite.cfg.CryptoKey, suite.publicFile)
	suite.loadCerts()

	repo := repository.NewRepository(&suite.cfg.StorageConfig, nil)
	suite.srv = service.NewService(repo, &suite.cfg.StorageConfig)

	ctx := context.Background()
	require.NoError(suite.T(), suite.Srv().SetGauge(ctx, "testGauge-1", domain.Gauge(1.0001)))
	require.NoError(suite.T(), suite.Srv().IncreaseCounter(ctx, "testCounter-1", domain.Counter(1)))

	_, err = suite.srv.SaveToFile(ctx)
	require.NoError(suite.T(), err)

	// clear OS ARGS
	// os.Args = make([]string, 0)

	suite.a = app.NewApp(suite.ctx, suite.stop,
		app.BuildMetadata{Version: "testing..", Date: time.Now().String(), Commit: ""},
		suite.cfg, zap.NewNop())

	go suite.a.Run()
}
func (suite *HandlerMemCryptoTestSuite) TearDownSuite() {
	require.NoError(suite.T(), os.RemoveAll(suite.T().TempDir()))
	suite.stop()
	suite.a.Stop()
}

func (suite *HandlerMemCryptoTestSuite) Srv() *service.Service {
	return suite.srv
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

			req, err := http.NewRequest(test.args.method, "http://"+suite.Cfg().Address+test.args.path, b)
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
