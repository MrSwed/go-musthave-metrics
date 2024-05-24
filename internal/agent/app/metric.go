package app

import (
	"fmt"

	"github.com/MrSwed/go-musthave-metrics/internal/agent/constant"
	myErr "github.com/MrSwed/go-musthave-metrics/internal/agent/error"
)

// Metric common metric structure for send
type Metric struct {
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	ID    string   `json:"id"`
	MType string   `json:"type"`
}

// NewMetric create new metric
func NewMetric(id, mType string) *Metric {
	return &Metric{
		ID:    id,
		MType: mType,
	}
}

// Set metric value
func (m *Metric) Set(v interface{}) (err error) {
	switch m.MType {
	case constant.GaugeType:
		var gv float64
		switch g := v.(type) {
		case float64:
			gv = g
		case int:
			gv = float64(g)
		case int32:
			gv = float64(g)
		case int64:
			gv = float64(g)
		case uint:
			gv = float64(g)
		case uint32:
			gv = float64(g)
		case uint64:
			gv = float64(g)
		default:
			return fmt.Errorf("%w %v", myErr.ErrBadGaugeValue, v)
		}
		m.Value = &gv
	case constant.CounterType:
		var cv int64
		switch c := v.(type) {
		case int64:
			cv = c
		case int:
			cv = int64(c)
		case int32:
			cv = int64(c)
		case uint64:
			cv = int64(c)
		case float32:
			cv = int64(c)
		case float64:
			cv = int64(c)
		default:
			return fmt.Errorf("%w %v", myErr.ErrBadCounterValue, v)
		}
		m.Delta = &cv
	default:
		err = fmt.Errorf("%w %s", myErr.ErrBadMetricType, m.MType)
		return
	}
	return nil
}
