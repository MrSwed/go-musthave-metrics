package server_test

import (
	"context"
	"crypto/rsa"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/MrSwed/go-musthave-metrics/internal/server/config"
	"github.com/MrSwed/go-musthave-metrics/internal/server/handler"
	"github.com/MrSwed/go-musthave-metrics/internal/server/repository"
	"github.com/MrSwed/go-musthave-metrics/internal/server/service"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type HandlerFileStoreTestSuite struct {
	suite.Suite
	ctx context.Context
	app http.Handler
	srv *service.Service
	cfg *config.Config
}

func (suite *HandlerFileStoreTestSuite) SetupSuite() {
	var (
		err    error
		logger *zap.Logger
	)
	suite.cfg = config.NewConfig()
	suite.cfg.FileStoreInterval = 0
	suite.cfg.FileStoragePath = filepath.Join(suite.T().TempDir(), "store.json")
	suite.ctx = context.Background()

	repo := repository.NewRepository(&suite.cfg.StorageConfig, nil)
	suite.srv = service.NewService(repo, &suite.cfg.StorageConfig)
	logger, err = zap.NewDevelopment()
	if err != nil {
		suite.Fail(err.Error())
	}

	suite.app = handler.NewHandler(suite.srv, &suite.cfg.WEB, logger).HTTPHandler()
}
func (suite *HandlerFileStoreTestSuite) TearDownSuite() {
	require.NoError(suite.T(), os.RemoveAll(suite.T().TempDir()))
}

func (suite *HandlerFileStoreTestSuite) App() http.Handler {
	return suite.app
}
func (suite *HandlerFileStoreTestSuite) Srv() *service.Service {
	return suite.srv
}
func (suite *HandlerFileStoreTestSuite) DBx() *sqlx.DB {
	return nil
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
}
