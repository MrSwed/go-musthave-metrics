package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMetric_set(t *testing.T) {
	tests := []struct {
		name    string
		mKey    string
		mType   string
		mValue  interface{}
		wantErr error
	}{
		{
			name:    "Set gauge float, Ok",
			mKey:    "someGauge",
			mType:   "gauge",
			mValue:  1.001,
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
			name:    "Set counter float, Ok",
			mKey:    "someCounter",
			mType:   "counter",
			mValue:  float64(1),
			wantErr: nil,
		},
		{
			name:    "Set counter string, Error",
			mKey:    "someCounter",
			mType:   "counter",
			mValue:  "string",
			wantErr: badCounterValue,
		},
		{
			name:    "Set gauge string, Error",
			mKey:    "someCounter",
			mType:   "gauge",
			mValue:  "string",
			wantErr: badGaugeValue,
		},
		{
			name:    "Set unknown type",
			mKey:    "someKey",
			mType:   "unknownType",
			mValue:  100,
			wantErr: badMetricType,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newMetric(tt.mKey, tt.mType)
			err := m.set(tt.mValue)
			if tt.wantErr == nil {
				require.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tt.wantErr)
			}
		})
	}
}
