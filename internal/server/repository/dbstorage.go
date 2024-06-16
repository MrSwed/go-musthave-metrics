package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"go-musthave-metrics/internal/server/constant"
	"go-musthave-metrics/internal/server/domain"
	myErr "go-musthave-metrics/internal/server/errors"

	"github.com/jmoiron/sqlx"
)

// DBStorageRepo is database repository
type DBStorageRepo struct {
	db *sqlx.DB
}

func NewDBStorageRepository(db *sqlx.DB) *DBStorageRepo {
	return &DBStorageRepo{
		db: db,
	}
}

// DBStorageGauge is gauge store
type DBStorageGauge struct {
	Name  string
	Value domain.Gauge
}

// DBStorageCounter is counter storage
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

// Ping check is db a life
func (r *DBStorageRepo) Ping(ctx context.Context) (err error) {
	if r.db == nil {
		return myErr.ErrNoDBConnected
	}
	err = retryFunc(func() (err error) {
		err = r.db.PingContext(ctx)
		return
	})
	return
}

// SetGauge save gauge to db
func (r *DBStorageRepo) SetGauge(ctx context.Context, k string, v domain.Gauge) (err error) {
	err = retryFunc(func() (err error) {
		_, err = r.db.ExecContext(ctx, `INSERT into `+constant.DBTableNameGauges+
			` (name, value) values ($1, $2) ON CONFLICT (name) DO UPDATE SET value = EXCLUDED.value`, k, v)
		return
	})
	return
}

// SetCounter save counter to db
func (r *DBStorageRepo) SetCounter(ctx context.Context, k string, v domain.Counter) (err error) {
	err = retryFunc(func() (err error) {
		_, err = r.db.ExecContext(ctx, `INSERT into `+constant.DBTableNameCounters+
			` (name, value) values ($1, $2) ON CONFLICT (name) DO UPDATE SET value = EXCLUDED.value`, k, v)
		return
	})
	return
}

// GetGauge get gauge from db
func (r *DBStorageRepo) GetGauge(ctx context.Context, k string) (v domain.Gauge, err error) {
	err = retryFunc(func() (err error) {
		err = r.db.GetContext(ctx, &v, `SELECT value FROM `+constant.DBTableNameGauges+
			` WHERE name = $1`, k)
		if errors.Is(err, sql.ErrNoRows) {
			err = myErr.ErrNotExist
		}
		return
	})
	return
}

// GetCounter get counter from db
func (r *DBStorageRepo) GetCounter(ctx context.Context, k string) (v domain.Counter, err error) {
	err = retryFunc(func() (err error) {
		err = r.db.GetContext(ctx, &v, `SELECT value FROM `+constant.DBTableNameCounters+` WHERE name = $1 LIMIT 1`, k)
		if errors.Is(err, sql.ErrNoRows) {
			err = myErr.ErrNotExist
		}
		return
	})
	return
}

// GetAllCounters get all counters from db
func (r *DBStorageRepo) GetAllCounters(ctx context.Context) (data domain.Counters, err error) {
	err = retryFunc(func() (err error) {
		var rows *sql.Rows
		if rows, err = r.db.QueryContext(ctx, `SELECT name, value FROM `+constant.DBTableNameCounters); err != nil {
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

// GetAllGauges get all gauges from db
func (r *DBStorageRepo) GetAllGauges(ctx context.Context) (data domain.Gauges, err error) {
	err = retryFunc(func() (err error) {
		var rows *sql.Rows
		if rows, err = r.db.QueryContext(ctx, `SELECT name, value FROM `+constant.DBTableNameGauges); err != nil {
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

// SetMetrics save several metrics to db
func (r *DBStorageRepo) SetMetrics(ctx context.Context, metrics []domain.Metric) (newMetrics []domain.Metric, err error) {
	newMetrics = make([]domain.Metric, len(metrics))
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
		if stmtG, err = tx.PreparexContext(ctx, "INSERT INTO "+constant.DBTableNameGauges+
			" (name, value) VALUES($1, $2) ON CONFLICT (name) DO UPDATE SET value = EXCLUDED.value"); err != nil {
			return
		}
		if stmtC, err = tx.PreparexContext(ctx, "INSERT INTO "+constant.DBTableNameCounters+" as c "+
			" (name, value) VALUES($1, $2) "+
			"ON CONFLICT (name) DO UPDATE SET value = c.value + EXCLUDED.value "+
			"RETURNING c.value"); err != nil {
			return
		}
		defer func() {
			err = errors.Join(err, stmtG.Close())
			err = errors.Join(err, stmtC.Close())
		}()

		for i, metric := range metrics {
			switch metric.MType {
			case constant.MetricTypeGauge:
				if _, err = stmtG.ExecContext(ctx, metric.ID, *metric.Value); err != nil {
					return
				}
			case constant.MetricTypeCounter:
				if err = stmtC.GetContext(ctx, metric.Delta, metric.ID, *metric.Delta); err != nil {
					return
				}
			}
			newMetrics[i] = metric
		}
		err = tx.Commit()
		return
	})
	return
}

// MemStore return memory store off all metrics
func (r *DBStorageRepo) MemStore(ctx context.Context) (m *MemStorageRepo, err error) {
	var (
		counters domain.Counters
		gauges   domain.Gauges
	)
	if counters, err = r.GetAllCounters(ctx); err != nil {
		return
	}
	if gauges, err = r.GetAllGauges(ctx); err != nil {
		return
	}
	m = &MemStorageRepo{
		MemStorageCounter: MemStorageCounter{Counter: counters},
		MemStorageGauge:   MemStorageGauge{Gauge: gauges},
	}

	return
}
