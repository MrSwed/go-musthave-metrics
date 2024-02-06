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
	return &DBStorageRepo{
		db: db,
	}
}

func (r *DBStorageRepo) Ping() error {
	if r.db == nil {
		return fmt.Errorf("no db connected")
	}
	return r.db.Ping()
}
