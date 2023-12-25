package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_getMetrics(t *testing.T) {

	tests := []struct {
		name string
		m    *metricsCollects
	}{
		{
			name: "Get metrics",
			m:    new(metricsCollects),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getMetrics(tt.m)
			assert.NotEmpty(t, tt.m)
		})
	}
}
