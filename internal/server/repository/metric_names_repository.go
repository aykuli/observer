package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	createMetricNamesTableQuery = `CREATE TABLE IF NOT EXISTS metric_names (
		 id SERIAL PRIMARY KEY,
		 name TEXT NOT NULL
	)`
	selectMetricNameQuery = `SELECT * FROM metric_names WHERE name=$1`
	createMetricQuery     = `INSERT INTO metric_names (name) VALUES ($1) RETURNING id, name`
	selectAllQuery        = `SELECT * FROM metric_names`
)

type MetricNamesRepository struct {
	client *pgxpool.Conn
}

type MetricName struct {
	ID   int
	Name string
}

func NewMetricNamesRepository(client *pgxpool.Conn) *MetricNamesRepository {
	return &MetricNamesRepository{client}
}

func (r *MetricNamesRepository) InitTable() error {
	if _, err := r.client.Exec(context.Background(), createMetricNamesTableQuery); err != nil {
		return err
	}

	return nil
}

func (r *MetricNamesRepository) FindByName(ctx context.Context, name string) (MetricName, error) {
	var metricName MetricName
	result := r.client.QueryRow(ctx, selectMetricNameQuery, name)
	if err := result.Scan(&metricName.ID, &metricName.Name); err != nil {
		row := r.client.QueryRow(ctx, createMetricQuery, name)
		err = row.Scan(&metricName.ID, &metricName.Name)
		if err != nil {
			return MetricName{}, err
		}

		return metricName, nil
	}

	return metricName, nil
}

func (r *MetricNamesRepository) SelectAll(ctx context.Context) ([]MetricName, error) {
	var metricNames []MetricName
	rows, err := r.client.Query(ctx, selectAllQuery)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var mName MetricName
		err = rows.Scan(&mName.ID, &mName.Name)
		if err != nil {
			return nil, err
		}
		metricNames = append(metricNames, mName)
	}

	return metricNames, nil
}

func (r *MetricNamesRepository) SelectByNames(ctx context.Context, names []string) ([]MetricName, error) {
	var metricNames []MetricName
	for _, name := range names {
		mtName, err := r.FindByName(ctx, name)
		if err != nil {
			return nil, err
		}

		metricNames = append(metricNames, mtName)
	}

	return metricNames, nil
}
