package main

import (
	"log"
	"net/http"

	"github.com/aykuli/observer/cmd/server/routers"
	"github.com/aykuli/observer/internal/server/config"
	"github.com/aykuli/observer/internal/server/logger"
	"github.com/aykuli/observer/internal/server/storage"
	"github.com/aykuli/observer/internal/server/storage/file"
	"github.com/aykuli/observer/internal/server/storage/postgres"
	"github.com/aykuli/observer/internal/server/storage/ram"
)

func main() {
	if err := logger.Initialize("Info"); err != nil {
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

	if config.Options.FileStoragePath != "" {
		return file.NewStorage(config.Options.FileStoragePath, config.Options.Restore, config.Options.StoreInterval)
	}

	return ram.NewStorage()
}
