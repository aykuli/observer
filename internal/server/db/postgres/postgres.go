package postgres

import (
	"context"
	"log"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/aykuli/observer/internal/server/config"
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

	return Instance, nil
}
