package repository

import (
	"context"
	"fmt"
	"testing"

	"go-musthave-metrics/internal/server/domain"
)

func BenchmarkMemStorageRepo_SetMetrics(b *testing.B) {
	r := NewMemRepository()
	const size = 10000
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
