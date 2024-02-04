package main

import (
	"log"
	"net/http"

	"github.com/aykuli/observer/cmd/server/handlers"
	"github.com/aykuli/observer/internal/storage"
)

func main() {
	memstorage := storage.MemStorage{
		GaugeMetrics:   map[string]float64{},
		CounterMetrics: map[string]int64{},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/update/", handlers.UpdateRuntime(memstorage))

	log.Fatal(http.ListenAndServe(`:8080`, mux))
}
