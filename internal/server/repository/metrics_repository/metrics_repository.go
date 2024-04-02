package metrics_repository

import (
	"context"

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
	//todo create constraint to metric that metric_id is a foreign_key to metric_names
	insertMetricQuery   = `INSERT INTO metrics (metric_id, type, value, delta) VALUES ($1, $2, $3, $4)`
	findByMetricIdQuery = `SELECT * FROM metrics WHERE metric_id=$1 ORDER BY created_at DESC`
)

type Repository struct {
	client *pgxpool.Conn
}

func NewRepository(client *pgxpool.Conn) *Repository {
	return &Repository{client}
}

func (r *Repository) InitTable() error {
	if _, err := r.client.Exec(context.Background(), createMetricsTableQuery); err != nil {
		return err
	}

	return nil
}

func (r *Repository) Insert(ctx context.Context, metricID int, metric models.Metric) error {
	var value *float64
	var delta *int64
	if metric.Value != nil {
		value = metric.Value
	}
	if metric.Delta != nil {
		delta = metric.Delta
	}

	if _, err := r.client.Exec(ctx, insertMetricQuery, metricID, metric.MType, value, delta); err != nil {
		return err
	}

	return nil
}

func (r *Repository) FindByMetricId(ctx context.Context, metricID int) (*models.Metric, error) {
	result := r.client.QueryRow(ctx, findByMetricIdQuery, metricID)
	var metric models.Metric0
	if err := result.Scan(&metric); err != nil {
		return nil, err
	}

	outMetric := &models.Metric{
		ID:    "",
		MType: metric.Type,
		Delta: nil,
		Value: nil,
	}

	if metric.Value.Valid {
		outMetric.Value = &metric.Value.Float64
	}

	if metric.Delta.Valid {
		outMetric.Delta = &metric.Delta.Int64
	}
	return outMetric, nil
}
