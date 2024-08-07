package postgres

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/aykuli/observer/internal/models"
	"github.com/aykuli/observer/internal/server/repository"
)

var pgOnce sync.Once

type DBStorage struct {
	instance *pgxpool.Pool
}

func NewStorage(dsn string) (*DBStorage, error) {
	ctx := context.Background()
	var s = DBStorage{}
	if err := s.createDBPool(ctx, dsn); err != nil {
		return &DBStorage{}, err
	}
	if err := s.createMetricsTable(ctx); err != nil {
		return &DBStorage{}, err
	}

	return &s, nil
}

func (s *DBStorage) createDBPool(ctx context.Context, dsn string) error {
	var resErr error
	pgOnce.Do(func() {
		pool, err := pgxpool.New(ctx, dsn)
		if err != nil {
			resErr = err
			return
		}

		s.instance = pool
	})

	if resErr != nil {
		return resErr
	}

	return nil
}

func (s *DBStorage) createMetricsTable(ctx context.Context) error {
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
