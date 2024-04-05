package repository

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/aykuli/observer/internal/server/models"
)

var (
	createMetricsTableQuery = `CREATE TABLE IF NOT EXISTS metrics (
	    id SERIAL PRIMARY KEY,
		metric_id INTEGER NOT NULL,
		type TEXT NOT NULL,
	    value FLOAT,
		delta INTEGER,
		created_at TIMESTAMPTZ NOT NULL DEFAULT  NOW()
	)`
	createIndexQuery    = `CREATE INDEX metric_id ON metric_names (id)`
	insertMetricQuery   = `INSERT INTO metrics (metric_id, type, value, delta) VALUES ($1, $2, $3, $4)`
	findByMetricIDQuery = `SELECT type, value, delta FROM metrics WHERE metric_id=$1 ORDER BY created_at DESC`
)

type MetricDB struct {
	ID       int
	MetricID int
	Type     string
	Delta    sql.NullInt64
	Value    sql.NullFloat64
}

type MetricsRepository struct {
	client *pgxpool.Conn
}

func NewMetricsRepository(client *pgxpool.Conn) *MetricsRepository {
	return &MetricsRepository{client}
}

func (r *MetricsRepository) InitTable() error {
	ctx := context.Background()
	if _, err := r.client.Exec(ctx, createMetricsTableQuery); err != nil {
		return err
	}

	if _, err := r.client.Exec(ctx, createIndexQuery); err != nil {
		return err
	}

	return nil
}

func (r *MetricsRepository) Insert(ctx context.Context, metricID int, metric models.Metric) error {
	if metric.MType == "gauge" {
		var value float64
		if metric.Value != nil {
			value = *metric.Value
		}

		if _, err := r.client.Exec(ctx, insertMetricQuery, metricID, metric.MType, value, nil); err != nil {
			return err
		}
	} else {
		var delta int64
		if metric.Delta != nil {
			delta = *metric.Delta
		}

		// get prev counter result
		result := r.client.QueryRow(ctx, findByMetricIDQuery, metricID)
		var metricValue sql.NullFloat64
		var metricDelta sql.NullInt64
		var metricType string
		var resultDelta = delta
		err := result.Scan(&metricType, &metricValue, &metricDelta)
		if err == nil {
			if metricDelta.Valid {
				resultDelta = delta + metricDelta.Int64
			}
		}

		if _, err := r.client.Exec(ctx, insertMetricQuery, metricID, metric.MType, nil, resultDelta); err != nil {
			return err
		}
	}

	return nil
}

func (r *MetricsRepository) InsertBatch(ctx context.Context, metrics []models.Metric, metricNames map[string]int) error {
	for _, mt := range metrics {
		err := r.Insert(ctx, metricNames[mt.ID], mt)

		if err != nil {
			return err
		}
	}

	return nil
}

func (r *MetricsRepository) FindByMetricName(ctx context.Context, metricName MetricName) (*models.Metric, error) {
	result := r.client.QueryRow(ctx, findByMetricIDQuery, metricName.ID)
	var metricValue sql.NullFloat64
	var metricDelta sql.NullInt64
	var metricType string
	if err := result.Scan(&metricType, &metricValue, &metricDelta); err != nil {
		return nil, err
	}

	outMetric := &models.Metric{
		ID:    metricName.Name,
		MType: metricType,
		Delta: nil,
		Value: nil,
	}

	if metricValue.Valid {
		outMetric.Value = &metricValue.Float64
	}
	if metricDelta.Valid {
		outMetric.Delta = &metricDelta.Int64
	}

	return outMetric, nil
}
