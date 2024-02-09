package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/MrSwed/go-musthave-metrics/internal/config"
	myErr "github.com/MrSwed/go-musthave-metrics/internal/errors"
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

type DBStorageGauge struct {
	Name  string
	Value float64
}

type DBStorageCounter struct {
	Name  string
	Value int64
}

// todo own types
//type DBStorageCounters map[string]int64
//type DBStorageGauges map[string]float64

func (r *DBStorageRepo) Ping() error {
	if r.db == nil {
		return fmt.Errorf("no db connected")
	}
	return r.db.Ping()
}

func (r *DBStorageRepo) SetGauge(k string, v float64) (err error) {
	_, err = r.db.Exec(`INSERT into `+config.DBTableNameGauges+
		` (name, value) values ($1, $2) ON CONFLICT (name) DO UPDATE SET value = EXCLUDED.value`, k, v)
	return
}

func (r *DBStorageRepo) SetCounter(k string, v int64) (err error) {
	_, err = r.db.Exec(`INSERT into `+config.DBTableNameCounters+
		` (name, value) values ($1, $2) ON CONFLICT (name) DO UPDATE SET value = EXCLUDED.value`, k, v)
	return
}

func (r *DBStorageRepo) GetGauge(k string) (v float64, err error) {
	err = r.db.Get(&v, `SELECT value FROM `+config.DBTableNameGauges+
		` WHERE name = $1`, k)
	if errors.Is(err, sql.ErrNoRows) {
		err = myErr.ErrNotExist
	}
	return
}

func (r *DBStorageRepo) GetCounter(k string) (v int64, err error) {
	err = r.db.Get(&v, `SELECT value FROM `+config.DBTableNameCounters+` WHERE name = $1 LIMIT 1`, k)
	if errors.Is(err, sql.ErrNoRows) {
		err = myErr.ErrNotExist
	}
	return
}

func (r *DBStorageRepo) GetAllCounters() (data map[string]int64, err error) {
	var rows *sql.Rows
	if rows, err = r.db.Query(`SELECT name, value FROM ` + config.DBTableNameCounters); err != nil {
		return
	}
	if err = rows.Err(); err != nil {
		return
	}
	defer rows.Close()
	data = make(map[string]int64)
	for rows.Next() {
		var item DBStorageCounter
		if err = rows.Scan(&item.Name, &item.Value); err != nil {
			return
		}
		data[item.Name] = item.Value
	}
	return
}

func (r *DBStorageRepo) GetAllGauges() (data map[string]float64, err error) {
	var rows *sql.Rows
	if rows, err = r.db.Query(`SELECT name, value FROM ` + config.DBTableNameGauges); err != nil {
		return
	}
	if err = rows.Err(); err != nil {
		return
	}
	defer rows.Close()
	data = make(map[string]float64)
	for rows.Next() {
		var item DBStorageGauge
		if err = rows.Scan(&item.Name, &item.Value); err != nil {
			return
		}
		data[item.Name] = item.Value
	}
	return
}
