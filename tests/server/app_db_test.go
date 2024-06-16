package server_test

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"testing"

	"go-musthave-metrics/internal/server/app"
	"go-musthave-metrics/internal/server/config"
	errM "go-musthave-metrics/internal/server/migrate"
	"go-musthave-metrics/internal/server/repository"
	"go-musthave-metrics/internal/server/service"

	"log"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
)

func CreatePostgresContainer(ctx context.Context) (*postgres.PostgresContainer, error) {
	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16.2-alpine3.19"),
		postgres.WithDatabase("test-db"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)

	if err != nil {
		return nil, err
	}

	return pgContainer, nil
}

type HandlerDBTestSuite struct {
	suite.Suite
	ctx    context.Context
	stop   context.CancelFunc
	srv    *service.Service
	cfg    *config.Config
	db     *sqlx.DB
	pgCont *postgres.PostgresContainer
}

func (suite *HandlerDBTestSuite) Srv() *service.Service {
	return suite.srv
}
func (suite *HandlerDBTestSuite) Cfg() *config.Config {
	return suite.cfg
}
func (suite *HandlerDBTestSuite) PublicKey() *rsa.PublicKey {
	return nil
}

func (suite *HandlerDBTestSuite) SetupSuite() {
	var (
		err error
	)
	suite.cfg = config.NewConfig()
	suite.ctx, suite.stop = context.WithCancel(context.Background())

	suite.pgCont, err = CreatePostgresContainer(suite.ctx)
	if err != nil {
		log.Fatal(err)
	}
	suite.cfg.DatabaseDSN, err = suite.pgCont.ConnectionString(suite.ctx, "sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	suite.db = func() *sqlx.DB {
		db, er := sqlx.Connect("pgx", suite.cfg.DatabaseDSN)
		if er != nil {
			log.Fatal(er)
		}
		return db
	}()

	_, err = errM.Migrate(suite.db.DB)
	switch {
	case err == nil:
	case errors.Is(err, migrate.ErrNoChange):
	default:
		log.Fatal(err)
	}

	suite.cfg.StorageConfig.FileStoragePath = filepath.Join(suite.T().TempDir(), fmt.Sprintf("store-data-%d.json", rand.Intn(200000)))
	suite.cfg.Address = net.JoinHostPort("localhost", fmt.Sprintf("%d", rand.Intn(200)+20000))
	suite.cfg.GRPCAddress = net.JoinHostPort("", fmt.Sprintf("%d", rand.Intn(200)+30000))
	suite.cfg.GRPCToken = "#GRPCSomeTokenString#"

	repo := repository.NewRepository(&suite.cfg.StorageConfig, suite.db)

	suite.srv = service.NewService(repo, &suite.cfg.StorageConfig)

	testData(suite)

	go app.RunApp(suite.ctx, suite.cfg, zap.NewNop(),
		app.BuildMetadata{Version: "test", Date: time.Now().Format(time.RFC3339), Commit: "test"})
	require.NoError(suite.T(), WaitHTTPPort(suite.ctx, suite))
	require.NoError(suite.T(), WaitGRPCPort(suite.ctx, suite))
}

func (suite *HandlerDBTestSuite) TearDownSuite() {
	if err := suite.pgCont.Terminate(suite.ctx); err != nil {
		log.Fatalf("error terminating postgres container: %s", err)
	}
	suite.stop()
	require.NoError(suite.T(), os.RemoveAll(suite.T().TempDir()))
}

func TestHandlersDB(t *testing.T) {
	suite.Run(t, new(HandlerDBTestSuite))
}

// TestMigrate
// migrate call at setup suite, so no test run without first migrate
// this is just for test cover
func (suite *HandlerDBTestSuite) TestMigrate() {
	testMigrate(suite, suite.db)
}

func (suite *HandlerDBTestSuite) TestGetMetric() {
	testGetMetric(suite)
}
func (suite *HandlerDBTestSuite) TestGetListMetrics() {
	testGetListMetrics(suite)
}
func (suite *HandlerDBTestSuite) TestGetMetricJson() {
	testGetMetricJSON(suite)
}
func (suite *HandlerDBTestSuite) TestUpdateMetric() {
	testUpdateMetric(suite)
}
func (suite *HandlerDBTestSuite) TestUpdateMetricJSON() {
	testUpdateMetricJSON(suite)
}
func (suite *HandlerDBTestSuite) TestUpdateMetrics() {
	testUpdateMetrics(suite)
}
func (suite *HandlerDBTestSuite) TestGzip() {
	testGzip(suite)
}
func (suite *HandlerDBTestSuite) TestHashKey() {
	testHashKey(suite)
}
func (suite *HandlerDBTestSuite) TestPing() {
	testPing(suite)
}

func (suite *HandlerDBTestSuite) TestGRPCGetMetric() {
	testGRPCGetMetric(suite)
}

func (suite *HandlerDBTestSuite) TestGRPCGetMetrics() {
	testGRPCGetMetrics(suite)
}

func (suite *HandlerDBTestSuite) TestGRPCSetMetric() {
	testGRPCSetMetric(suite)
}

func (suite *HandlerDBTestSuite) TestGRPCSetMetrics() {
	testGRPCSetMetrics(suite)
}
