package server_test

import (
	"context"
	"crypto/rsa"
	"fmt"
	"go-musthave-metrics/internal/server/app"
	"go-musthave-metrics/internal/server/repository"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go-musthave-metrics/internal/server/config"
	"go-musthave-metrics/internal/server/service"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type HandlerMemTestSuite struct {
	suite.Suite
	ctx  context.Context
	stop context.CancelFunc
	srv  *service.Service
	cfg  *config.Config
	a    *app.App
}

func (suite *HandlerMemTestSuite) Srv() *service.Service {
	return suite.srv
}
func (suite *HandlerMemTestSuite) Cfg() *config.Config {
	return suite.cfg
}
func (suite *HandlerMemTestSuite) PublicKey() *rsa.PublicKey {
	return nil
}

func (suite *HandlerMemTestSuite) SetupSuite() {

	suite.cfg = config.NewConfig()
	suite.ctx, suite.stop = context.WithCancel(context.Background())
	suite.cfg.StorageConfig.FileStoragePath = filepath.Join(suite.T().TempDir(), fmt.Sprintf("store-data-%d.json", rand.Intn(200000)))
	suite.cfg.Address = net.JoinHostPort("localhost", fmt.Sprintf("%d", rand.Intn(200)+20000))
	suite.cfg.GRPCAddress = net.JoinHostPort("", fmt.Sprintf("%d", rand.Intn(200)+30000))

	repo := repository.NewRepository(&suite.cfg.StorageConfig, nil)
	suite.srv = service.NewService(repo, &suite.cfg.StorageConfig)

	testData(suite)

	suite.a = app.NewApp(suite.ctx, suite.stop,
		app.BuildMetadata{Version: "testing..", Date: time.Now().String(), Commit: ""},
		suite.cfg, zap.NewNop())

	go suite.a.Run()
}

func (suite *HandlerMemTestSuite) TearDownSuite() {
	require.NoError(suite.T(), os.RemoveAll(suite.T().TempDir()))
	suite.stop()
	suite.a.Stop()
}

func TestHandlersMem(t *testing.T) {
	suite.Run(t, new(HandlerMemTestSuite))
}

func (suite *HandlerMemTestSuite) TestGetMetric() {
	testGetMetric(suite)
}
func (suite *HandlerMemTestSuite) TestGetListMetrics() {
	testGetListMetrics(suite)
}
func (suite *HandlerMemTestSuite) TestGetMetricJSON() {
	testGetMetricJSON(suite)
}
func (suite *HandlerMemTestSuite) TestUpdateMetric() {
	testUpdateMetric(suite)
}
func (suite *HandlerMemTestSuite) TestUpdateMetricJson() {
	testUpdateMetricJSON(suite)
}
func (suite *HandlerMemTestSuite) TestUpdateMetrics() {
	testUpdateMetrics(suite)
}
func (suite *HandlerMemTestSuite) TestGzip() {
	testGzip(suite)
}
func (suite *HandlerMemTestSuite) TestHashKey() {
	testHashKey(suite)
}
func (suite *HandlerMemTestSuite) TestPing() {
	testPing(suite)
}

func (suite *HandlerMemTestSuite) TestSaveToFile() {
	testSaveToFile(suite, suite.ctx)
}

func (suite *HandlerMemTestSuite) TestGRPCGetMetric() {
	testGRPCGetMetric(suite)
}

func (suite *HandlerMemTestSuite) TestGRPCGetMetrics() {
	testGRPCGetMetrics(suite)
}

func (suite *HandlerMemTestSuite) TestGRPCSetMetric() {
	testGRPCSetMetric(suite)
}

func (suite *HandlerMemTestSuite) TestGRPCSetMetrics() {
	testGRPCSetMetrics(suite)
}
