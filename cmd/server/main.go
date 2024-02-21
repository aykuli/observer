package main

import (
	"log"
	"net/http"

	"github.com/aykuli/observer/cmd/server/routers"
	"github.com/aykuli/observer/internal/server/config"
	"github.com/aykuli/observer/internal/server/logger"
	"github.com/aykuli/observer/internal/storage"
)

func main() {
	if err := logger.Initialize("info"); err != nil {
		log.Fatal(err)
	}

	memStorage := storage.MemStorageInit

	log.Fatal(http.ListenAndServe(config.ListenAddr, logger.WithLogging(routers.MetricsRouter(&memStorage))))
}
