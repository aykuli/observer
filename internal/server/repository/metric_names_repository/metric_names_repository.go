package metric_names_repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	createMetricNamesTableQuery = `CREATE TABLE IF NOT EXISTS metric_names (
		 id SERIAL PRIMARY KEY,
		 name TEXT NOT NULL
	)`
	selectMetricNameQuery = `SELECT id FROM metric_names WHERE name=$1`
	createMetricQuery     = `INSERT INTO metric_names (name) VALUES ($1) RETURNING id`
	selectAllQuery        = `SELECT * FROM metric_names`
)

type Repository struct {
	client *pgxpool.Conn
}

type MetricName struct {
	Id   int
	Name string
}

func NewRepository(client *pgxpool.Conn) *Repository {
	return &Repository{client}
}

func (r *Repository) InitTable() error {
	if _, err := r.client.Exec(context.Background(), createMetricNamesTableQuery); err != nil {
		return err
	}

	return nil

}

func (r *Repository) GetID(ctx context.Context, name string) (int, error) {
	var metricID int
	result := r.client.QueryRow(ctx, selectMetricNameQuery, name)
	if err := result.Scan(&metricID); err != nil {
		row := r.client.QueryRow(ctx, createMetricQuery, name)
		err = row.Scan(&metricID)
		if err != nil {
			return metricID, err
		}

		return metricID, nil
	}

	return metricID, nil
}

func (r *Repository) SelectAll(ctx context.Context) ([]MetricName, error) {
	var metricNames []MetricName
	rows, err := r.client.Query(ctx, selectAllQuery)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var mName MetricName
		err = rows.Scan(&mName.Id, &mName.Name)
		if err != nil {
			return nil, err
		}
		metricNames = append(metricNames, mName)
	}

	return metricNames, nil
}
