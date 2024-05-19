package server_test

import (
	"context"
	"crypto/rsa"
	"errors"
	"net/http"
	"testing"

	"github.com/MrSwed/go-musthave-metrics/internal/server/config"
	"github.com/MrSwed/go-musthave-metrics/internal/server/handler"
	errM "github.com/MrSwed/go-musthave-metrics/internal/server/migrate"
	"github.com/MrSwed/go-musthave-metrics/internal/server/repository"
	"github.com/MrSwed/go-musthave-metrics/internal/server/service"
	"github.com/go-chi/chi/v5"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"

	"log"
	"time"
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
	app    http.Handler
	srv    *service.Service
	cfg    *config.Config
	db     *sqlx.DB
	pgCont *postgres.PostgresContainer
}

func (suite *HandlerDBTestSuite) App() http.Handler {
	return suite.app
}
func (suite *HandlerDBTestSuite) Srv() *service.Service {
	return suite.srv
}

func (suite *HandlerDBTestSuite) DBx() *sqlx.DB {
	return suite.db
}
func (suite *HandlerDBTestSuite) Cfg() *config.Config {
	return suite.cfg
}
func (suite *HandlerDBTestSuite) PublicKey() *rsa.PublicKey {
	return nil
}

func (suite *HandlerDBTestSuite) SetupSuite() {
	var (
		err    error
		logger *zap.Logger
	)
	suite.cfg = config.NewConfig()
	suite.ctx = context.Background()
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

	repo := repository.NewRepository(&suite.cfg.StorageConfig, suite.db)

	suite.srv = service.NewService(repo, &suite.cfg.StorageConfig)
	logger, err = zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	suite.app = handler.NewHandler(chi.NewRouter(), suite.srv, &suite.cfg.WEB, logger).Handler()
}

func (suite *HandlerDBTestSuite) TearDownSuite() {
	if err := suite.pgCont.Terminate(suite.ctx); err != nil {
		log.Fatalf("error terminating postgres container: %s", err)
	}
}

func TestHandlersDB(t *testing.T) {
	suite.Run(t, new(HandlerDBTestSuite))
}

// TestMigrate
// migrate call at setup suite, so no test run without first migrate
// this is just for test cover
func (suite *HandlerDBTestSuite) TestMigrate() {
	testMigrate(suite)
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
