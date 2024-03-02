package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/aykuli/observer/internal/server/models"
)

var (
	createMetricsTableQuery = `CREATE TABLE IF NOT EXISTS metrics (
	    id SERIAL PRIMARY KEY,
		metric_id INTEGER NOT NULL,
		type TEXT NOT NULL,
	    value FLOAT,
		delta TEXT,
		created_at TIMESTAMPTZ NOT NULL DEFAULT  NOW(),
		CONSTRAINT metric_id_fkey FOREIGN KEY(metric_id) REFERENCES metric_names(id)
	)`
	insertMetricQuery   = `INSERT INTO metrics (metric_id, type, value, delta) VALUES (@metric_id, @type, @value, @delta)`
	findByMetricIDQuery = `SELECT type, value, delta FROM metrics WHERE metric_id=$1 ORDER BY created_at DESC`
)

type MetricDB struct {
	ID       int
	MetricID int
	Type     string
	Delta    sql.NullString
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

	return nil
}

func (r *MetricsRepository) Insert(ctx context.Context, metricID int, metric models.Metric) error {
	if metric.MType == "gauge" {
		var value float64
		if metric.Value != nil {
			value = *metric.Value
		}
		args := pgx.NamedArgs{"metric_id": metricID, "type": metric.MType, "value": value, "delta": nil}
		if _, err := r.client.Exec(ctx, insertMetricQuery, args); err != nil {
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
		var metricDelta sql.NullString
		var metricType string
		var resultDelta = delta
		err := result.Scan(&metricType, &metricValue, &metricDelta)
		if err == nil {
			if metricDelta.Valid {
				parsedDelta, err := strconv.ParseInt(metricDelta.String, 10, 64)
				if err != nil {
					return err
				}
				resultDelta = delta + parsedDelta
			}
		}

		args := pgx.NamedArgs{"metric_id": metricID, "type": metric.MType, "value": nil, "delta": fmt.Sprint(resultDelta)}
		if _, err := r.client.Exec(ctx, insertMetricQuery, args); err != nil {
			return err
		}
	}

	return nil
}

func (r *MetricsRepository) InsertBatch(ctx context.Context, metrics []models.Metric, metricNames map[string]int) error {
	var errs []error
	for _, mt := range metrics {
		err := r.Insert(ctx, metricNames[mt.ID], mt)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (r *MetricsRepository) FindByMetricName(ctx context.Context, metricName MetricName) (*models.Metric, error) {
	result := r.client.QueryRow(ctx, findByMetricIDQuery, metricName.ID)
	var metricValue sql.NullFloat64
	var metricDelta sql.NullString
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
		parsedDelta, err := strconv.ParseInt(metricDelta.String, 10, 64)
		if err != nil {
			return nil, err
		}
		outMetric.Delta = &parsedDelta
	}

	return outMetric, nil
}
