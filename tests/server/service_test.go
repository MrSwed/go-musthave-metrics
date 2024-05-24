package server

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	"github.com/MrSwed/go-musthave-metrics/internal/server/config"
	"github.com/MrSwed/go-musthave-metrics/internal/server/domain"
	"github.com/MrSwed/go-musthave-metrics/internal/server/repository"
	"github.com/MrSwed/go-musthave-metrics/internal/server/service"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ServiceTestSuite struct {
	suite.Suite
	ctx context.Context
	srv *service.Service
	cfg *config.Config
}

func (suite *ServiceTestSuite) SetupSuite() {
	suite.cfg = config.NewConfig()
	suite.ctx = context.Background()
	repo := repository.NewRepository(&suite.cfg.StorageConfig, nil)
	suite.srv = service.NewService(repo, &suite.cfg.StorageConfig)

}

func TestServices(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}

func (suite *ServiceTestSuite) TestRestoreFromFile() {
	t := suite.T()
	t.Run("Restore from file", func(t *testing.T) {
		_, err := suite.srv.RestoreFromFile(suite.ctx)
		require.NoError(t, err)
	})
}

func (suite *ServiceTestSuite) TestSaveToFile() {
	t := suite.T()
	testCounter := domain.Counter(1)
	testGauge := domain.Gauge(1.0001)
	testGaugeName := fmt.Sprintf("testGauge%d", rand.Int())
	testCounterName := fmt.Sprintf("testCounter%d", rand.Int())
	// save some values
	ctx := context.Background()
	_ = suite.srv.SetGauge(ctx, testGaugeName, testGauge)
	_ = suite.srv.IncreaseCounter(ctx, testCounterName, testCounter)

	t.Run("Save file", func(t *testing.T) {
		_, err := suite.srv.SaveToFile(suite.ctx)
		require.NoError(t, err)
	})
}
