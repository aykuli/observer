package postgres

import (
	"context"
	"log"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/aykuli/observer/internal/server/config"
	"github.com/aykuli/observer/internal/server/repository/metric_names_repository"
	"github.com/aykuli/observer/internal/server/repository/metrics_repository"
)

var (
	Instance *pgxpool.Pool
	pgOnce   sync.Once
)

func CreateDBPool() (*pgxpool.Pool, error) {
	pgOnce.Do(func() {
		pool, err := pgxpool.New(context.Background(), config.Options.DatabaseDsn)
		if err != nil {
			log.Print(err)
		}

		Instance = pool
	})

	conn, err := Instance.Acquire(context.Background())
	if err != nil {
		log.Print(err)

	}
	defer conn.Release()

	metricNamesRepo := metric_names_repository.NewRepository(conn)
	if err = metricNamesRepo.InitTable(); err != nil {
		log.Print(err)
	}

	metricsRepo := metrics_repository.NewRepository(conn)
	if err = metricsRepo.InitTable(); err != nil {
		log.Print(err)
	}

	return Instance, nil
}
