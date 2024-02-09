package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/aykuli/observer/cmd/server/routers"
	"github.com/aykuli/observer/internal/storage"
)

func main() {
	parseFlags()
	memStorage := storage.MemStorage{
		GaugeMetrics:   map[string]float64{},
		CounterMetrics: map[string]int64{},
	}

	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%v", addr.Host, addr.Port), routers.MetricsRouter(&memStorage)))
}
