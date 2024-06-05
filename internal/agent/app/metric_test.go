package app

import (
	"testing"

	myErr "go-musthave-metrics/internal/agent/error"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetric_set(t *testing.T) {
	tests := []struct {
		mValue  interface{}
		wantErr error
		name    string
		mKey    string
		mType   string
	}{
		{
			name:    "Set gauge float, Ok",
			mKey:    "someGauge",
			mType:   "gauge",
			mValue:  1.001,
			wantErr: nil,
		},
		{
			name:    "Set gauge int, Ok",
			mKey:    "someGauge",
			mType:   "gauge",
			mValue:  1,
			wantErr: nil,
		},
		{
			name:    "Set gauge int32, Ok",
			mKey:    "someGauge",
			mType:   "gauge",
			mValue:  int32(1),
			wantErr: nil,
		}, {
			name:    "Set gauge uint, Ok",
			mKey:    "someGauge",
			mType:   "gauge",
			mValue:  uint(1),
			wantErr: nil,
		},
		{
			name:    "Set gauge int64, Ok",
			mKey:    "someGauge",
			mType:   "gauge",
			mValue:  int64(1),
			wantErr: nil,
		},
		{
			name:    "Set gauge uint64, Ok",
			mKey:    "someGauge",
			mType:   "gauge",
			mValue:  uint64(1),
			wantErr: nil,
		},
		{
			name:    "Set counter, Ok",
			mKey:    "someCounter",
			mType:   "counter",
			mValue:  1,
			wantErr: nil,
		},
		{
			name:    "Set counter float64, Ok",
			mKey:    "someCounter",
			mType:   "counter",
			mValue:  float64(1),
			wantErr: nil,
		},
		{
			name:    "Set counter float32, Ok",
			mKey:    "someCounter",
			mType:   "counter",
			mValue:  float32(1),
			wantErr: nil,
		},
		{
			name:    "Set counter uint64, Ok",
			mKey:    "someCounter",
			mType:   "counter",
			mValue:  uint64(1),
			wantErr: nil,
		},
		{
			name:    "Set counter int32, Ok",
			mKey:    "someCounter",
			mType:   "counter",
			mValue:  int32(1),
			wantErr: nil,
		},
		{
			name:    "Set counter string, Error",
			mKey:    "someCounter",
			mType:   "counter",
			mValue:  "string",
			wantErr: myErr.ErrBadCounterValue,
		},
		{
			name:    "Set gauge string, Error",
			mKey:    "someCounter",
			mType:   "gauge",
			mValue:  "string",
			wantErr: myErr.ErrBadGaugeValue,
		},
		{
			name:    "Set unknown type",
			mKey:    "someKey",
			mType:   "unknownType",
			mValue:  100,
			wantErr: myErr.ErrBadMetricType,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMetric(tt.mKey, tt.mType)
			err := m.Set(tt.mValue)
			if tt.wantErr == nil {
				require.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tt.wantErr)
			}
		})
	}
}
