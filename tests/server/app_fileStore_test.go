package server_test

import (
	"context"
	"crypto/rsa"
	"fmt"
	"go-musthave-metrics/internal/server/app"
	"go-musthave-metrics/internal/server/config"
	"go-musthave-metrics/internal/server/domain"
	"go-musthave-metrics/internal/server/repository"
	"go-musthave-metrics/internal/server/service"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type HandlerFileStoreTestSuite struct {
	suite.Suite
	ctx  context.Context
	stop context.CancelFunc
	srv  *service.Service
	cfg  *config.Config
}

func (suite *HandlerFileStoreTestSuite) SetupSuite() {
	suite.cfg = config.NewConfig()
	suite.cfg.FileStoreInterval = 0
	suite.cfg.FileStoragePath = filepath.Join(suite.T().TempDir(), "store.json")
	suite.ctx, suite.stop = context.WithCancel(context.Background())

	suite.cfg.Address = net.JoinHostPort("localhost", fmt.Sprintf("%d", rand.Intn(200)+20000))
	suite.cfg.GRPCAddress = net.JoinHostPort("", fmt.Sprintf("%d", rand.Intn(200)+30000))

	repo := repository.NewRepository(&suite.cfg.StorageConfig, nil)
	suite.srv = service.NewService(repo, &suite.cfg.StorageConfig)

	testData(suite)

	go app.RunApp(suite.ctx, suite.cfg, zap.NewNop(),
		app.BuildMetadata{Version: "test", Date: time.Now().Format(time.RFC3339), Commit: "test"})

	require.NoError(suite.T(), WaitHTTPPort(suite.ctx, suite))
	require.NoError(suite.T(), WaitGRPCPort(suite.ctx, suite))
}

func (suite *HandlerFileStoreTestSuite) TearDownSuite() {
	require.NoError(suite.T(), os.RemoveAll(suite.T().TempDir()))
	suite.stop()
}

func (suite *HandlerFileStoreTestSuite) Srv() *service.Service {
	return suite.srv
}
func (suite *HandlerFileStoreTestSuite) Cfg() *config.Config {
	return suite.cfg
}
func (suite *HandlerFileStoreTestSuite) PublicKey() *rsa.PublicKey {
	return nil
}

func TestHandlersFileStoreTest(t *testing.T) {
	suite.Run(t, new(HandlerFileStoreTestSuite))
}

func (suite *HandlerFileStoreTestSuite) TestUpdateMetric() {
	testUpdateMetric(suite)
}

func (suite *HandlerFileStoreTestSuite) TestUpdateMetricJson() {
	testUpdateMetricJSON(suite)
}

func (suite *HandlerFileStoreTestSuite) TestUpdateMetrics() {
	testUpdateMetrics(suite)
}

func (suite *HandlerFileStoreTestSuite) TestRestoreFromFile() {
	t := suite.T()
	t.Run("Restore from file", func(t *testing.T) {
		_, err := suite.srv.RestoreFromFile(suite.ctx)
		require.NoError(t, err)
	})
	t.Run("Restore from file Err", func(t *testing.T) {
		sp := suite.Cfg().FileStoragePath
		suite.Cfg().FileStoragePath = ""
		defer func() {
			suite.Cfg().FileStoragePath = sp
		}()
		_, err := suite.srv.RestoreFromFile(suite.ctx)
		assert.EqualError(t, err, "no storage file provided")
	})
}

func (suite *HandlerMemTestSuite) TestSaveToFile() {

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
		_, err := suite.Srv().SaveToFile(suite.ctx)
		require.NoError(t, err)
	})

	t.Run("Save file Error", func(t *testing.T) {
		sp := suite.Cfg().FileStoragePath
		suite.Cfg().FileStoragePath = ""
		defer func() {
			suite.Cfg().FileStoragePath = sp
		}()
		_, err := suite.Srv().SaveToFile(suite.ctx)
		assert.EqualError(t, err, "no storage file provided")
	})
}
