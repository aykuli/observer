package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/aykuli/observer/cmd/server/routers"
	"github.com/aykuli/observer/internal/server/config"
	"github.com/aykuli/observer/internal/server/logger"
	"github.com/aykuli/observer/internal/server/storage"
	"github.com/aykuli/observer/internal/server/storage/local"
	"github.com/aykuli/observer/internal/server/storage/postgres"
)

func main() {
	if err := logger.Initialize("debug"); err != nil {
		log.Print(err)
	}

	memStorage, err := initStorage()
	if err != nil {
		log.Print(err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		fmt.Println("Gracefully shutting down observer application and closing database")
		memStorage.Close()
	}()

	if err = http.ListenAndServe(config.Options.Address, logger.WithLogging(routers.MetricsRouter(memStorage))); err != nil {
		c <- os.Interrupt
	}
}

func initStorage() (storage.Storage, error) {
	if config.Options.DatabaseDsn != "" {
		return postgres.NewStorage(config.Options.DatabaseDsn)
	}

	return local.NewStorage(config.Options)
}
