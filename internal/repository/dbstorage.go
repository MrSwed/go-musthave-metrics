package repository

import (
	"fmt"
	"github.com/jmoiron/sqlx"
)

type DBStorage interface {
	Ping() error
}

type DBStorageRepo struct {
	db *sqlx.DB
}

func NewDBStorageRepository(db *sqlx.DB) *DBStorageRepo {
	if db == nil {
		return nil
	}
	return &DBStorageRepo{
		db: db,
	}
}

func (r *DBStorageRepo) Ping() error {
	if r.db == nil {
		return fmt.Errorf("no db connect")
	}
	return r.db.Ping()
}
