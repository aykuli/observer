package main

import (
	"fmt"
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
			log.Fatal(err)
		}
	}
	fmt.Printf("options: %+v\n\n", config.Options)

	if err := logger.Initialize("info"); err != nil {
		log.Print(err)
	}

	memStorage := storage.MemStorage{
		GaugeMetrics:   storage.GaugeMetrics{},
		CounterMetrics: storage.CounterMetrics{},
	}

	if err := memStorage.Load(); err != nil {
		log.Print(err)
	}

	saveToFile := config.Options.DatabaseDsn == "" && config.Options.SaveMetrics && config.Options.StoreInterval > 0
	if saveToFile {
		go memStorage.SaveMetricsPeriodically()
	}

	if err := http.ListenAndServe(config.Options.Address, logger.WithLogging(routers.MetricsRouter(&memStorage))); err != nil {
		log.Fatal(err)
	}
}
