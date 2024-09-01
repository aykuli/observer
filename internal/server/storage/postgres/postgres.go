// Package postgres provides handler to store metrics in database.
package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/aykuli/observer/internal/models"
	"github.com/aykuli/observer/internal/server/repository"
)

type DBStorage struct {
	instance *pgxpool.Pool
}

func NewStorage(dsn string) (*DBStorage, error) {
	ctx := context.Background()
	var s DBStorage
	tryCount := 0
	createConn := func() error {
		word := "try"
		if tryCount > 0 {
			word = "retry"
		}
		fmt.Printf("%s to connect to database, probe %d\n", word, tryCount)
		tryCount++

		pool, err := pgxpool.New(ctx, dsn)
		if err != nil {
			return fmt.Errorf("could not connect to database: %v", err)
		}

		err = pool.Ping(ctx)
		if err != nil {
			return fmt.Errorf("could not connect to database: %v", err)
		}

		fmt.Printf("  | -- connected to database %s\n", dsn)
		s.instance = pool

		return nil
	}
	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.MaxElapsedTime = 12 * time.Second
	if err := backoff.Retry(createConn, expBackoff); err != nil {
		return nil, fmt.Errorf("\nfailed to connect to database after retrying %d times: %v", tryCount, err)
	}
	if err := s.createMetricsTableIfNotExists(ctx); err != nil {
		return &s, err
	}

	return &s, nil
}

func (s *DBStorage) createMetricsTableIfNotExists(ctx context.Context) error {
	conn, err := s.instance.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	metricsRepo := repository.NewMetricsRepository(conn)
	if err := metricsRepo.InitTable(ctx); err != nil {
		return newDBError(err)
	}

	return nil
}

func (s *DBStorage) Close() {
	s.instance.Close()
}

func (s *DBStorage) Ping(ctx context.Context) error {
	return s.instance.Ping(ctx)
}

func (s *DBStorage) GetMetrics(ctx context.Context) (string, error) {
	conn, err := s.instance.Acquire(ctx)
	if err != nil {
		return "", newDBError(err)
	}
	defer conn.Release()

	metricsRepo := repository.NewMetricsRepository(conn)

	metrics, err := metricsRepo.SelectAllValues(ctx)
	if err != nil {
		return "", newDBError(err)
	}

	return s.parseMetrics(metrics), nil
}

func (s *DBStorage) parseMetrics(metrics []models.Metric) string {
	var pair = make([]string, len(metrics))
	for i, m := range metrics {
		var valueStr string
		switch m.MType {
		case "gauge":
			valueStr = fmt.Sprintf("%f", *m.Value)
		case "counter":
			valueStr = fmt.Sprintf("%d", *m.Delta)
		}
		pair[i] = "   " + m.ID + valueStr
	}

	return strings.Join(pair, ",\n")
}

func (s *DBStorage) ReadMetric(ctx context.Context, mName, mType string) (*models.Metric, error) {
	conn, err := s.instance.Acquire(ctx)
	if err != nil {
		return nil, newDBError(err)
	}
	defer conn.Release()

	metricsRepo := repository.NewMetricsRepository(conn)
	metric, err := metricsRepo.FindByNameAndType(ctx, mName, mType)
	if err != nil {
		return nil, newDBError(err)
	}
	return metric, nil
}

func (s *DBStorage) SaveMetric(ctx context.Context, metric models.Metric) (*models.Metric, error) {
	conn, err := s.instance.Acquire(ctx)
	if err != nil {
		return nil, newDBError(err)
	}
	defer conn.Release()

	metricsRepo := repository.NewMetricsRepository(conn)
	tx, err := conn.Begin(ctx)
	if err != nil {
		return nil, newDBError(err)
	}

	outMetric, err := metricsRepo.Save(ctx, tx, metric)
	if err != nil {
		if err = tx.Rollback(ctx); err != nil {
			return nil, newDBError(err)
		}
		return nil, newDBError(err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, newDBError(err)
	}

	return outMetric, nil
}

func (s *DBStorage) SaveBatch(ctx context.Context, metrics []models.Metric) ([]models.Metric, error) {
	conn, err := s.instance.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	metricsRepo := repository.NewMetricsRepository(conn)

	tx, err := conn.Begin(ctx)
	if err != nil {
		return nil, err
	}

	outMetrics, err := metricsRepo.SaveBatch(ctx, tx, metrics)
	if err != nil {
		if err = tx.Rollback(ctx); err != nil {
			return nil, newDBError(err)
		}
		return nil, newDBError(err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, newDBError(err)
	}

	return outMetrics, nil
}
