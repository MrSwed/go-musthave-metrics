package server_test

import (
	"bytes"
	"context"
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"errors"
	"go-musthave-metrics/internal/server/config"
	"go-musthave-metrics/internal/server/domain"
	errM "go-musthave-metrics/internal/server/migrate"
	"go-musthave-metrics/internal/server/service"
	"net"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

type HandlerTestSuite interface {
	Srv() *service.Service
	T() *testing.T
	Cfg() *config.Config
	PublicKey() *rsa.PublicKey
}

const waitPortInterval = 100 * time.Millisecond
const waitPortConnTimeout = 50 * time.Millisecond

func testData(suite HandlerTestSuite) {
	ctx := context.Background()
	require.NoError(suite.T(), suite.Srv().SetGauge(ctx, "testGauge-1", domain.Gauge(1.0001)))
	require.NoError(suite.T(), suite.Srv().IncreaseCounter(ctx, "testCounter-1", domain.Counter(1)))

	_, err := suite.Srv().SaveToFile(ctx)
	require.NoError(suite.T(), err)
}

func testMigrate(suite HandlerTestSuite, db *sqlx.DB) {
	t := suite.T()
	t.Run("Migrate", func(t *testing.T) {
		_, err := errM.Migrate(db.DB)
		switch {
		case errors.Is(err, migrate.ErrNoChange):
		default:
			require.NoError(t, err)
		}
	})
}

func maybeCryptBody(bodyBuf *bytes.Buffer, publicKey *rsa.PublicKey) {
	if publicKey != nil {
		cipherBody, err := rsa.EncryptOAEP(sha256.New(), crand.Reader, publicKey, bodyBuf.Bytes(), nil)
		bodyBuf.Reset()
		if err != nil {
			bodyBuf.WriteString(err.Error())
			return
		}
		bodyBuf.Write(cipherBody)
	}
}

func WaitHTTPPort(ctx context.Context, suite HandlerTestSuite) error {
	if suite.Cfg().Address == "" {
		return nil
	}
	ticker := time.NewTicker(waitPortInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			conn, _ := net.DialTimeout("tcp", suite.Cfg().Address, waitPortConnTimeout)
			if conn != nil {
				_ = conn.Close()
				return nil
			}
		}
	}
}

func WaitGRPCPort(ctx context.Context, suite HandlerTestSuite) error {
	if suite.Cfg().GRPCAddress == "" {
		return nil
	}
	ticker := time.NewTicker(waitPortInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			conn, _ := net.DialTimeout("tcp", suite.Cfg().GRPCAddress, waitPortConnTimeout)
			if conn != nil {
				_ = conn.Close()
				return nil
			}
		}
	}
}
