package errors

import (
	"errors"
)

var (
	ErrNotExist      = errors.New("does not exist")
	ErrNoDBConnected = errors.New("no DB connected")
	ErrNotMemMode    = errors.New("no MemStore connected")
)
