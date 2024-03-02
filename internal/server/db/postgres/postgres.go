package postgres

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/aykuli/observer/internal/server/config"
	"github.com/aykuli/observer/internal/server/repository"
)

var (
	Instance *pgxpool.Pool
	pgOnce   sync.Once
)

type DBError struct {
	Name string
	Err  error
}

func (dbErr *DBError) Error() string {
	return fmt.Sprintf("[%s] %v", dbErr.Name, dbErr.Err)
}

func NewDBError(err error) error {
	var pgErr *pgconn.PgError
	name := "postgres"
	if errors.As(err, &pgErr) {
		if pgerrcode.IsConnectionException(pgErr.Code) {
			name = "Connection Exception"
		}

		if pgerrcode.IsDataException(pgErr.Code) {
			name = "Data Exception"
		}

	}

	return &DBError{
		Name: name,
		Err:  err,
	}
}

func CreateDBPool() (*pgxpool.Pool, error) {
	var resErr error
	pgOnce.Do(func() {
		pool, err := pgxpool.New(context.Background(), config.Options.DatabaseDsn)
		if err != nil {
			resErr = err
			return
		}

		Instance = pool
	})

	if resErr != nil {
		return nil, NewDBError(resErr)
	}

	conn, err := Instance.Acquire(context.Background())
	if err != nil {
		return nil, NewDBError(err)
	}
	defer conn.Release()

	var errs []error
	metricNamesRepo := repository.NewMetricNamesRepository(conn)
	if err = metricNamesRepo.InitTable(); err != nil {
		errs = append(errs, NewDBError(err))
	}

	metricsRepo := repository.NewMetricsRepository(conn)
	if err = metricsRepo.InitTable(); err != nil {
		errs = append(errs, NewDBError(err))
	}

	return Instance, errors.Join(errs...)
}
