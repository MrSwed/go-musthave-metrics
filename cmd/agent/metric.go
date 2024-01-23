package main

import (
	"errors"
	"fmt"
)

var (
	badGaugeValue   = errors.New("bad gauge value")
	badCounterValue = errors.New("bad counter value")
	badMetricType   = errors.New("unknown metric type")
)

type metric struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

func newMetric(id, mType string) *metric {
	return &metric{
		ID:    id,
		MType: mType,
	}
}

func (m *metric) set(v interface{}) (err error) {
	switch m.MType {
	case gaugeType:
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
			return fmt.Errorf("%w %v", badGaugeValue, v)
		}
		m.Value = &gv
	case counterType:
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
			return fmt.Errorf("%w %v", badCounterValue, v)
		}
		m.Delta = &cv
	default:
		err = fmt.Errorf("%w %s", badMetricType, m.MType)
		return
	}
	return nil
}
