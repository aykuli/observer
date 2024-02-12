package main

import (
	"log"
	"net/http"

	"github.com/aykuli/observer/cmd/server/routers"
	"github.com/aykuli/observer/internal/server/config"
	"github.com/aykuli/observer/internal/storage"
)

func main() {
	memStorage := storage.MemStorageInit

	log.Fatal(http.ListenAndServe(config.ListenAddr, routers.MetricsRouter(&memStorage)))
}
