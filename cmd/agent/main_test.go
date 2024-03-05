package main

import (
	"github.com/MrSwed/go-musthave-metrics/internal/agent/app"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_getMetrics(t *testing.T) {

	tests := []struct {
		name string
		m    *app.MetricsCollects
	}{
		{
			name: "Get metrics",
			m:    new(app.MetricsCollects),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.m.GetMetrics()
			assert.NotEmpty(t, tt.m)
		})
	}
}
