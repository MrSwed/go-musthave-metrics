package errors

import (
	"testing"

	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
)

func TestIsPQClass08Error(t *testing.T) {
	tests := []struct {
		err     error
		name    string
		wantYes bool
	}{
		{
			name:    "Nil",
			err:     nil,
			wantYes: false,
		},
		{
			name: "ConnectionException",
			err: &pq.Error{
				Code: pgerrcode.ConnectionException,
			},
			wantYes: true,
		},
		{
			name: "ConnectionDoesNotExist",
			err: &pq.Error{
				Code: pgerrcode.ConnectionDoesNotExist,
			},
			wantYes: true,
		},
		{
			name: "ConnectionFailure",
			err: &pq.Error{
				Code: pgerrcode.ConnectionFailure,
			},
			wantYes: true,
		},
		{
			name: "SQLClientUnableToEstablishSQLConnection",
			err: &pq.Error{
				Code: pgerrcode.SQLClientUnableToEstablishSQLConnection,
			},
			wantYes: true,
		},
		{
			name: "SQLServerRejectedEstablishmentOfSQLConnection",
			err: &pq.Error{
				Code: pgerrcode.SQLServerRejectedEstablishmentOfSQLConnection,
			},
			wantYes: true,
		},
		{
			name: "TransactionResolutionUnknown",
			err: &pq.Error{
				Code: pgerrcode.TransactionResolutionUnknown,
			},
			wantYes: true,
		},
		{
			name: "ProtocolViolation",
			err: &pq.Error{
				Code: pgerrcode.ProtocolViolation,
			},
			wantYes: true,
		},
		{
			name: "SQLStatementNotYetComplete",
			err: &pq.Error{
				Code: pgerrcode.SQLStatementNotYetComplete,
			},
			wantYes: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotYes := IsPQClass08Error(tt.err); gotYes != tt.wantYes {
				t.Errorf("IsPQClass08Error() = %v, want %v", gotYes, tt.wantYes)
			}
		})
	}
}
