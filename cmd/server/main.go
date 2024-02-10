package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/aykuli/observer/cmd/server/routers"
	"github.com/aykuli/observer/internal/storage"
)

var listenAddr string

func main() {
	parseFlags()
	parseEnvVars()

	memStorage := storage.MemStorage{
		GaugeMetrics:   map[string]float64{},
		CounterMetrics: map[string]int64{},
	}
	fmt.Println(listenAddr)
	log.Fatal(http.ListenAndServe(listenAddr, routers.MetricsRouter(&memStorage)))
}
