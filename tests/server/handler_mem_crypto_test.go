package server_test

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"net/http"
	"os"
	"testing"

	"github.com/MrSwed/go-musthave-metrics/internal/server/config"
	"github.com/MrSwed/go-musthave-metrics/internal/server/handler"
	"github.com/MrSwed/go-musthave-metrics/internal/server/repository"
	"github.com/MrSwed/go-musthave-metrics/internal/server/service"
	testHelpers "github.com/MrSwed/go-musthave-metrics/tests"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type HandlerMemCryptoTestSuite struct {
	suite.Suite
	ctx        context.Context
	app        http.Handler
	srv        *service.Service
	cfg        *config.Config
	publicFile string
	publicKey  *rsa.PublicKey
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

	suite.app = handler.NewHandler(chi.NewRouter(), suite.srv, &suite.cfg.WEB, logger).Handler()
}
func (suite *HandlerMemCryptoTestSuite) TearDownSuite() {
	if suite.cfg.CryptoKey != "" {
		_ = os.Remove(suite.cfg.CryptoKey)
		_ = os.Remove(suite.publicFile)
	}
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
