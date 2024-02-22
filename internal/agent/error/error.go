package error

import (
	"errors"
	"fmt"
	"runtime"
)

var (
	ErrBadGaugeValue   = errors.New("bad gauge value")
	ErrBadCounterValue = errors.New("bad counter value")
	ErrBadMetricType   = errors.New("unknown metric type")
)

// ErrWrap wrap error with debug info: line and file name where it happened
func ErrWrap(err error) error {
	_, fn, line, _ := runtime.Caller(1)
	return fmt.Errorf("%w at %s:%d", err, fn, line)
}
