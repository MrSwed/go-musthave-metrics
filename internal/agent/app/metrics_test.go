package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getMetrics(t *testing.T) {

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
