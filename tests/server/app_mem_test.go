package server_test

import (
	"context"
	"crypto/rsa"
	"net/http"
	"testing"

	"github.com/MrSwed/go-musthave-metrics/internal/server/config"
	"github.com/MrSwed/go-musthave-metrics/internal/server/handler"
	"github.com/MrSwed/go-musthave-metrics/internal/server/repository"
	"github.com/MrSwed/go-musthave-metrics/internal/server/service"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type HandlerMemTestSuite struct {
	suite.Suite
	ctx context.Context
	app http.Handler
	srv *service.Service
	cfg *config.Config
}

func (suite *HandlerMemTestSuite) SetupSuite() {
	var (
		err    error
		logger *zap.Logger
	)
	suite.cfg = config.NewConfig()
	suite.ctx = context.Background()

	repo := repository.NewRepository(&suite.cfg.StorageConfig, nil)

	suite.srv = service.NewService(repo, &suite.cfg.StorageConfig)
	logger, err = zap.NewDevelopment()
	if err != nil {
		suite.Fail(err.Error())
	}

	suite.app = handler.NewHandler(chi.NewRouter(), suite.srv, &suite.cfg.WEB, logger).Handler()
}

func (suite *HandlerMemTestSuite) App() http.Handler {
	return suite.app
}
func (suite *HandlerMemTestSuite) Srv() *service.Service {
	return suite.srv
}
func (suite *HandlerMemTestSuite) DBx() *sqlx.DB {
	return nil
}
func (suite *HandlerMemTestSuite) Cfg() *config.Config {
	return suite.cfg
}
func (suite *HandlerMemTestSuite) PublicKey() *rsa.PublicKey {
	return nil
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
