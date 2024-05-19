package repository

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/aykuli/observer/internal/models"
)

var (
	createMetricsTableQuery = `CREATE TABLE IF NOT EXISTS metrics (
		name VARCHAR NOT NULL,
		type TEXT NOT NULL,
	  value FLOAT,
		delta BIGINT)`
	selectAllLastMetricsQuery    = `SELECT name, type, value, delta FROM metrics ORDER BY name`
	findByMetricNameAndTypeQuery = `SELECT value, delta FROM metrics WHERE name=@name AND type=@type`

	updateGaugeQuery          = `UPDATE metrics SET value = @value WHERE name=@name AND type='gauge' RETURNING value`
	updateCounterQuery        = `UPDATE metrics SET delta = delta + @delta WHERE name=@name AND type='counter' RETURNING delta`
	insertMetricQuery         = `INSERT INTO metrics (name, type, value, delta) VALUES (@name, @type, @value, @delta) RETURNING value, delta`
	checkMetricExistanceQuery = `SELECT count(*) FROM metrics WHERE name=@name AND type=@type`
)

type MetricDB struct {
	Name  string
	Type  string
	Delta sql.NullString
	Value sql.NullFloat64
}

type MetricsRepository struct {
	conn *pgxpool.Conn
}

func NewMetricsRepository(client *pgxpool.Conn) *MetricsRepository {
	return &MetricsRepository{client}
}

func (r *MetricsRepository) InitTable(ctx context.Context) error {
	if _, err := r.conn.Exec(ctx, createMetricsTableQuery); err != nil {
		return err
	}

	return nil
}

func (r *MetricsRepository) SelectAllValues(ctx context.Context) ([]*models.Metric, error) {
	var metrics []*models.Metric

	result, err := r.conn.Query(ctx, selectAllLastMetricsQuery)
	if err != nil {
		return nil, err
	}
	for result.Next() {
		var m = &models.Metric{}
		var value sql.NullFloat64
		var delta sql.NullInt64

		if err = result.Scan(&m.ID, &m.MType, &value, &delta); err != nil {
			return nil, err
		}
		if value.Valid {
			v := value.Float64
			m.Value = &v
		}
		if delta.Valid {
			d := delta.Int64
			m.Delta = &d
		}
		metrics = append(metrics, m)
	}
	return metrics, nil
}

func (r *MetricsRepository) FindByNameAndType(ctx context.Context, mName, mType string) (*models.Metric, error) {
	outMt := models.Metric{ID: mName, MType: mType}

	args := pgx.NamedArgs{"name": mName, "type": mType}
	result := r.conn.QueryRow(ctx, findByMetricNameAndTypeQuery, args)
	var metricValue sql.NullFloat64
	var metricDelta sql.NullInt64
	if err := result.Scan(&metricValue, &metricDelta); err != nil {
		return nil, err
	}

	var value float64
	var delta int64
	if metricValue.Valid {
		value = metricValue.Float64
		outMt.Value = &value
	}
	if metricDelta.Valid {
		delta = metricDelta.Int64
		outMt.Delta = &delta
	}

	return &outMt, nil
}

// Save update metric if it exists else insert it
func (r *MetricsRepository) Save(ctx context.Context, metric models.Metric) (*models.Metric, error) {
	tx, err := r.conn.Begin(ctx)
	if err != nil {
		return nil, err
	}

	outMt, err := r.save(ctx, metric)
	if err != nil {
		if err = tx.Rollback(ctx); err != nil {
			return nil, err
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	return outMt, nil
}

func (r *MetricsRepository) save(ctx context.Context, metric models.Metric) (*models.Metric, error) {
	var outMt *models.Metric
	var exist int
	result := r.conn.QueryRow(ctx, checkMetricExistanceQuery, pgx.NamedArgs{"name": metric.ID, "type": metric.MType})
	err := result.Scan(&exist)
	if err != nil {
		return nil, err
	}
	if exist == 0 {
		outMt, err = r.insert(ctx, metric)
		if err != nil {
			return nil, err
		}
	} else {
		outMt, err = r.update(ctx, metric)
		if err != nil {
			return nil, err
		}
	}

	return outMt, nil
}

func (r *MetricsRepository) insert(ctx context.Context, metric models.Metric) (*models.Metric, error) {
	var outMt = metric
	var args pgx.NamedArgs
	var value float64
	var delta int64

	if metric.MType == "gauge" {
		value = *metric.Value
		args = pgx.NamedArgs{"name": metric.ID, "type": "gauge", "value": value, "delta": nil}
	} else if metric.MType == "counter" {
		delta = *metric.Delta
		args = pgx.NamedArgs{"name": metric.ID, "type": "counter", "value": nil, "delta": delta}

	} else {
		return nil, pgx.ErrNoRows
	}

	result := r.conn.QueryRow(ctx, insertMetricQuery, args)

	var valueSQL sql.NullFloat64
	var deltaSQL sql.NullInt64
	if err := result.Scan(&valueSQL, &deltaSQL); err != nil {
		return nil, err
	}

	if valueSQL.Valid {
		outMt.Value = &valueSQL.Float64
	}
	if deltaSQL.Valid {
		outMt.Delta = &deltaSQL.Int64
	}

	return &outMt, nil
}

func (r *MetricsRepository) update(ctx context.Context, metric models.Metric) (*models.Metric, error) {
	outMt := models.Metric{ID: metric.ID, MType: metric.MType}

	if metric.MType == "gauge" {
		args := pgx.NamedArgs{"name": metric.ID, "type": "gauge", "value": &metric.Value, "delta": nil}
		if _, err := r.conn.Exec(ctx, updateGaugeQuery, args); err != nil {
			return &outMt, err
		}
		outMt.Value = metric.Value
		return &outMt, nil
	} else if metric.MType == "counter" {
		result := r.conn.QueryRow(ctx, updateCounterQuery, pgx.NamedArgs{"name": metric.ID, "delta": &metric.Delta})
		var delta int64
		if err := result.Scan(&delta); err != nil {
			return &outMt, err

		}

		outMt.Delta = &delta
		return &outMt, nil
	}

	return &outMt, pgx.ErrNoRows
}

func (r *MetricsRepository) SaveBatch(ctx context.Context, metrics []models.Metric) ([]models.Metric, error) {
	var outMts []models.Metric

	tx, err := r.conn.Begin(ctx)
	if err != nil {
		return nil, err
	}

	for _, mt := range metrics {
		outMt, err := r.save(ctx, mt)
		if err != nil {
			if err = tx.Rollback(ctx); err != nil {
				return nil, err
			}
			return nil, err
		}
		outMts = append(outMts, *outMt)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	return outMts, nil
}
