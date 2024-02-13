package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/MrSwed/go-musthave-metrics/internal/constant"
	"github.com/MrSwed/go-musthave-metrics/internal/domain"
	myErr "github.com/MrSwed/go-musthave-metrics/internal/errors"

	"github.com/jmoiron/sqlx"
)

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
	Value domain.Gauge
}

type DBStorageCounter struct {
	Name  string
	Value domain.Counter
}

func retryFunc(fn func() error) (err error) {
	for i := 0; i <= len(constant.Backoff); i++ {
		err = fn()
		if err == nil || !myErr.IsPQClass08Error(err) {
			break
		}
		if i < len(constant.Backoff) {
			time.Sleep(constant.Backoff[i])
		}
	}
	return
}

func (r *DBStorageRepo) Ping() (err error) {
	if r.db == nil {
		return fmt.Errorf("no db connected")
	}
	err = retryFunc(func() (err error) {
		err = r.db.Ping()
		return
	})
	return
}

func (r *DBStorageRepo) SetGauge(k string, v domain.Gauge) (err error) {
	err = retryFunc(func() (err error) {
		_, err = r.db.Exec(`INSERT into `+constant.DBTableNameGauges+
			` (name, value) values ($1, $2) ON CONFLICT (name) DO UPDATE SET value = EXCLUDED.value`, k, v)
		return
	})
	return
}

func (r *DBStorageRepo) SetCounter(k string, v domain.Counter) (err error) {
	err = retryFunc(func() (err error) {
		_, err = r.db.Exec(`INSERT into `+constant.DBTableNameCounters+
			` (name, value) values ($1, $2) ON CONFLICT (name) DO UPDATE SET value = EXCLUDED.value`, k, v)
		return
	})
	return
}

func (r *DBStorageRepo) GetGauge(k string) (v domain.Gauge, err error) {
	err = retryFunc(func() (err error) {
		err = r.db.Get(&v, `SELECT value FROM `+constant.DBTableNameGauges+
			` WHERE name = $1`, k)
		if errors.Is(err, sql.ErrNoRows) {
			err = myErr.ErrNotExist
		}
		return
	})
	return
}

func (r *DBStorageRepo) GetCounter(k string) (v domain.Counter, err error) {
	err = retryFunc(func() (err error) {
		err = r.db.Get(&v, `SELECT value FROM `+constant.DBTableNameCounters+` WHERE name = $1 LIMIT 1`, k)
		if errors.Is(err, sql.ErrNoRows) {
			err = myErr.ErrNotExist
		}
		return
	})
	return
}

func (r *DBStorageRepo) GetAllCounters() (data domain.Counters, err error) {
	err = retryFunc(func() (err error) {
		var rows *sql.Rows
		if rows, err = r.db.Query(`SELECT name, value FROM ` + constant.DBTableNameCounters); err != nil {
			return
		}
		if err = rows.Err(); err != nil {
			return
		}
		defer func(rows *sql.Rows) {
			err = rows.Close()
		}(rows)
		data = make(domain.Counters)
		for rows.Next() {
			var item DBStorageCounter
			if err = rows.Scan(&item.Name, &item.Value); err != nil {
				return
			}
			data[item.Name] = item.Value
		}
		return
	})
	return
}

func (r *DBStorageRepo) GetAllGauges() (data domain.Gauges, err error) {
	err = retryFunc(func() (err error) {
		var rows *sql.Rows
		if rows, err = r.db.Query(`SELECT name, value FROM ` + constant.DBTableNameGauges); err != nil {
			return
		}
		if err = rows.Err(); err != nil {
			return
		}
		defer func(rows *sql.Rows) {
			err = rows.Close()
		}(rows)
		data = make(domain.Gauges)
		for rows.Next() {
			var item DBStorageGauge
			if err = rows.Scan(&item.Name, &item.Value); err != nil {
				return
			}
			data[item.Name] = item.Value
		}
		return
	})
	return
}

func (r *DBStorageRepo) SetMetrics(metrics []domain.Metric) (newMetrics []domain.Metric, err error) {
	err = retryFunc(func() (err error) {
		var tx *sqlx.Tx
		tx, err = r.db.Beginx()
		if err != nil {
			return
		}
		defer func() {
			rErr := tx.Rollback()
			if rErr != nil && !errors.Is(rErr, sql.ErrTxDone) {
				err = errors.Join(err, rErr)
			}
		}()
		var stmtG, stmtC *sqlx.Stmt
		if stmtG, err = tx.Preparex("INSERT INTO " + constant.DBTableNameGauges +
			" (name, value) VALUES($1, $2) ON CONFLICT (name) DO UPDATE SET value = EXCLUDED.value"); err != nil {
			return
		}
		if stmtC, err = tx.Preparex("INSERT INTO " + constant.DBTableNameCounters + " as c " +
			" (name, value) VALUES($1, $2) " +
			"ON CONFLICT (name) DO UPDATE SET value = c.value + EXCLUDED.value " +
			"RETURNING c.value"); err != nil {
			return
		}
		defer func() {
			err = errors.Join(err, stmtG.Close())
			err = errors.Join(err, stmtC.Close())
		}()

		for _, metric := range metrics {
			switch metric.MType {
			case constant.MetricTypeGauge:
				if _, err = stmtG.Exec(metric.ID, *metric.Value); err != nil {
					return
				}
				newMetrics = append(newMetrics, metric)
			case constant.MetricTypeCounter:
				newMetric := metric
				if err = stmtC.Get(newMetric.Delta, metric.ID, *metric.Delta); err != nil {
					return
				}
				newMetrics = append(newMetrics, newMetric)
			}
		}
		err = tx.Commit()
		return
	})
	return
}

func (r *DBStorageRepo) MemStore() (m *MemStorageRepo, err error) {
	var (
		counters domain.Counters
		gauges   domain.Gauges
	)
	if counters, err = r.GetAllCounters(); err != nil {
		return
	}
	if gauges, err = r.GetAllGauges(); err != nil {
		return
	}
	m = &MemStorageRepo{
		MemStorageCounter: MemStorageCounter{Counter: counters},
		MemStorageGauge:   MemStorageGauge{Gauge: gauges},
	}

	return
}
