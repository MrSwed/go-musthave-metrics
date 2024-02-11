package errors

import (
	"errors"
	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
)

var (
	ErrNotExist      = errors.New("does not exist")
	ErrNoDBConnected = errors.New("no DB connected")
	ErrNotMemMode    = errors.New("no MemStore connected")
	PQError          = &pq.Error{}
)

func IsPQClass08Error(err error) (yes bool) {
	if err == nil {
		return
	}

	if e, ok := err.(*pq.Error); ok {
		return e.Code == pgerrcode.ConnectionException ||
			e.Code == pgerrcode.ConnectionDoesNotExist ||
			e.Code == pgerrcode.ConnectionFailure ||
			e.Code == pgerrcode.SQLClientUnableToEstablishSQLConnection ||
			e.Code == pgerrcode.SQLServerRejectedEstablishmentOfSQLConnection ||
			e.Code == pgerrcode.TransactionResolutionUnknown ||
			e.Code == pgerrcode.ProtocolViolation
	}
	return
}
