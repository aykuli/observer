package main

import (
	"log"
	"net/http"

	"github.com/aykuli/observer/cmd/server/routers"
	"github.com/aykuli/observer/internal/server/config"
	"github.com/aykuli/observer/internal/server/db/postgres"
	"github.com/aykuli/observer/internal/server/logger"
	"github.com/aykuli/observer/internal/server/storage"
)

func main() {
	if config.Options.DatabaseDsn != "" {
		if _, err := postgres.CreateDBPool(); err != nil {
			log.Print(err)
		}
	}

	if err := logger.Initialize("Info"); err != nil {
		log.Print(err)
	}

	memStorage := storage.MemStorage{
		GaugeMetrics:   storage.GaugeMetrics{},
		CounterMetrics: storage.CounterMetrics{},
	}

	workWithFile := config.Options.DatabaseDsn == "" && config.Options.FileStoragePath != ""

	if workWithFile {
		if config.Options.Restore {
			if err := memStorage.Load(); err != nil {
				log.Print(err)
			}
		}

		if config.Options.StoreInterval > 0 {
			go memStorage.SaveMetricsPeriodically()
		}
	}

	if err := http.ListenAndServe(config.Options.Address, logger.WithLogging(routers.MetricsRouter(&memStorage))); err != nil {
		log.Fatal(err)
	}
}
