package repository

import (
	"context"
	"fmt"
	"log"
	"testing"

	"go-musthave-metrics/internal/server/config"
	"go-musthave-metrics/internal/server/domain"

	"github.com/jmoiron/sqlx"
)

func NewConfigGetTest() (c *config.Config) {
	c = &config.Config{
		StorageConfig: config.StorageConfig{
			FileStoragePath: "",
			StorageRestore:  false,
		},
	}
	err := c.ParseEnv()
	if err != nil {
		log.Fatal(err)
	}
	c.CleanSchemes()

	if c.DatabaseDSN != "" {
		if dbTest, err = sqlx.Open("postgres", c.DatabaseDSN); err != nil {
			log.Fatal(err)
		}
	}
	return
}

var (
	_      = NewConfigGetTest()
	dbTest *sqlx.DB
)

func BenchmarkDbStorageRepo_SetMetrics(b *testing.B) {
	if dbTest == nil {
		fmt.Println("DatabaseDSN required")
		return
	}
	r := NewDBStorageRepository(dbTest)
	const size = 100
	metrics := make([]domain.Metric, size)
	for i := 0; i < size; i += 2 {
		metrics[i] = domain.Metric{
			ID:    fmt.Sprintf("testCount%d", i),
			MType: "counter",
			Delta: &[]domain.Counter{domain.Counter(i) + 10}[0],
		}
		metrics[i+1] = domain.Metric{
			ID:    fmt.Sprintf("testGauge%d", i+1),
			MType: "gauge",
			Value: &[]domain.Gauge{domain.Gauge(i+1) + 110}[0],
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		nM, _ := r.SetMetrics(context.Background(), metrics)
		_ = len(nM)
	}
}
