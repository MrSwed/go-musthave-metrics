package server_test

import (
	"bytes"
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"go-musthave-metrics/internal/server/handler/rest"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"testing"

	"go-musthave-metrics/internal/server/app"
	"go-musthave-metrics/internal/server/config"
	errM "go-musthave-metrics/internal/server/migrate"
	"go-musthave-metrics/internal/server/repository"
	"go-musthave-metrics/internal/server/service"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

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

	suite.app = rest.NewHandler(suite.srv, suite.cfg, logger).Handler()
}

func (suite *HandlerDBTestSuite) TearDownSuite() {
	if err := suite.pgCont.Terminate(suite.ctx); err != nil {
		log.Fatalf("error terminating postgres container: %s", err)
	}
	require.NoError(suite.T(), os.RemoveAll(suite.T().TempDir()))
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

func (suite *HandlerDBTestSuite) TestApp() {
	t := suite.T()

	type fields struct {
		ctx   context.Context
		stop  context.CancelFunc
		cfg   *config.Config
		build app.BuildMetadata
	}
	tests := []struct {
		name             string
		fields           fields
		wantStrings      []string
		doNotWantStrings []string
	}{
		{
			name: "Server app run. default",
			fields: func() fields {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				return fields{
					ctx:  ctx,
					stop: cancel,
					cfg: func() *config.Config {
						cfg := suite.cfg
						cfg.Address = ":" + strconv.Itoa(rand.Intn(65000)+1000)
						// cfg.FileStoragePath = filepath.Join(t.TempDir(), fmt.Sprintf("metrict-db-%d.json", rand.Int()))
						return cfg
					}(),
					build: app.BuildMetadata{
						Version: "1.0-testing",
						Date:    "24.05.24",
						Commit:  "444333",
					},
				}
			}(),
			wantStrings: []string{
				`"Init app"`,
				`"Build version":"1.0-testing"`,
				`"Build date":"24.05.24"`,
				`"Build commit":"444333"`,
				`Start server`,
				`http server started`,
				`grpc server started`,
				`Shutting down server gracefully`,
				`Store save on interval finished`,
				`Storage saved`,
				`Server stopped`,
			},
			doNotWantStrings: []string{
				`"error"`,
			},
		},
		{
			name: "Server app run. port busy",
			fields: func() fields {
				cfg := config.NewConfig()
				cfg.Address = ":" + strconv.Itoa(rand.Intn(65000)+1000)

				portUse, err := net.Listen("tcp", cfg.Address)
				require.NoError(t, err)

				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				return fields{
					ctx: ctx,
					stop: func() {
						_ = portUse.Close()
						cancel()
					},
					cfg: cfg,
					build: app.BuildMetadata{
						Version: "1.0-testing",
						Date:    "24.05.24",
						Commit:  "444333",
					},
				}
			}(),
			wantStrings: []string{
				`"Init app"`,
				`"Build version":"1.0-testing"`,
				`"Build date":"24.05.24"`,
				`"Build commit":"444333"`,
				`Start server`,
				`Shutting down server gracefully`,
				`"error"`,
				`listen tcp`,
				`bind: address already in use`,
			},
			doNotWantStrings: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, stop := signal.NotifyContext(tt.fields.ctx, os.Interrupt, syscall.SIGTERM)
			defer stop()
			defer tt.fields.stop()

			tt.fields.cfg = tt.fields.cfg.CleanSchemes()

			var buf bytes.Buffer
			logger := zap.New(func(pipeTo io.Writer) zapcore.Core {
				return zapcore.NewCore(
					zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
					zap.CombineWriteSyncers(os.Stderr, zapcore.AddSync(pipeTo)),
					zapcore.InfoLevel,
				)
			}(&buf))
			// flag.

			appHandler := app.NewApp(ctx, stop, tt.fields.build, tt.fields.cfg, logger)

			appHandler.Run()
			appHandler.Stop()

			t.Log(buf.String())
			for i := 0; i < len(tt.wantStrings); i++ {
				assert.Contains(t, buf.String(), tt.wantStrings[i], fmt.Sprintf("%s is expected at log out", tt.wantStrings[i]))
			}
			for i := 0; i < len(tt.doNotWantStrings); i++ {
				assert.NotContains(t, buf.String(), tt.doNotWantStrings[i], fmt.Sprintf("%s is not expected at log out", tt.doNotWantStrings[i]))
			}
		})
	}
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

func (suite *HandlerDBTestSuite) TestSaveToFile() {
	testSaveToFile(suite, suite.ctx)
}
