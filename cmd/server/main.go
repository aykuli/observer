package main

import (
	"log"
	"net/http"

	"github.com/aykuli/observer/cmd/server/routers"
	"github.com/aykuli/observer/internal/server/config"
	"github.com/aykuli/observer/internal/server/logger"
	"github.com/aykuli/observer/internal/server/storage"
	"github.com/aykuli/observer/internal/server/storage/local"
	"github.com/aykuli/observer/internal/server/storage/postgres"
)

func main() {
	if err := logger.Initialize("info"); err != nil {
		log.Print(err)
	}

	memStorage, err := initStorage()
	if err != nil {
		log.Print(err)
	}

	if err = http.ListenAndServe(config.Options.Address, logger.WithLogging(routers.MetricsRouter(memStorage))); err != nil {
		log.Fatal(err)
	}
}

func initStorage() (storage.Storage, error) {
	if config.Options.DatabaseDsn != "" {
		return postgres.NewStorage(config.Options.DatabaseDsn)
	}

	return local.NewStorage(config.Options)
}
