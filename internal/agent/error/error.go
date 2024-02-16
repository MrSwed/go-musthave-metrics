package error

import "errors"

var (
	ErrBadGaugeValue   = errors.New("bad gauge value")
	ErrBadCounterValue = errors.New("bad counter value")
	ErrBadMetricType   = errors.New("unknown metric type")
)
