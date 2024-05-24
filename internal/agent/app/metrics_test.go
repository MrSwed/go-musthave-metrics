package app

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/MrSwed/go-musthave-metrics/internal/agent/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsCollects_getMetrics(t *testing.T) {

	tests := []struct {
		m    *MetricsCollects
		name string
	}{
		{
			name: "Get metrics",
			m:    new(MetricsCollects),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.m.GetMetrics()
			assert.NotEmpty(t, tt.m)
		})
	}
}

func TestMetricsCollects_GopMetrics(t *testing.T) {
	tests := []struct {
		m    *MetricsCollects
		name string
	}{
		{
			name: "Get gop metrics",
			m:    new(MetricsCollects),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.m.GetGopMetrics()
			require.NoError(t, err)
			assert.NotEmpty(t, tt.m)
		})
	}
}

func TestMetricsCollects_ListMetrics(t *testing.T) {
	c := config.NewConfig()
	m := NewMetricsCollects(c)

	t.Run("Get ListMetrics", func(t *testing.T) {
		metrics, err := m.ListMetrics()
		require.NoError(t, err)
		assert.NotEmpty(t, metrics)
	})
}

func TestMetricsCollects_SendMetrics(t *testing.T) {
	c := config.NewConfig()
	m := NewMetricsCollects(c)

	t.Run("Send Metrics", func(t *testing.T) {
		n, err := m.SendMetrics(context.TODO())
		var allowErr *net.OpError
		if assert.Error(t, err) && !errors.As(err, &allowErr) {
			require.NoError(t, err)
		}
		assert.NotEmpty(t, n)
	})
}
