package main

import (
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/aykuli/observer/cmd/agent/handlers"
	"github.com/aykuli/observer/internal/storage"
)

const (
	pollInterval   = 2
	reportInterval = 10
)

func main() {
	urlBase := "http://localhost:8080"
	request := resty.New().R()

	memStorage := storage.MemStorage{GaugeMetrics: map[string]float64{}, CounterMetrics: map[string]int64{}}
	collectTicker := time.NewTicker(pollInterval * time.Second)
	collectQuit := make(chan struct{})

	sendTicker := time.NewTicker(reportInterval * time.Second)
	sendQuit := make(chan struct{})

	i := 0
	for i < 5 {
		i++
		for {
			select {
			case <-collectTicker.C:
				storage.GetStats(&memStorage)
			case <-sendTicker.C:
				handlers.SendPost(request, urlBase, memStorage)
			case <-sendQuit:
				sendTicker.Stop()
				return
			case <-collectQuit:
				collectTicker.Stop()
				return
			}
		}
	}

	defer collectTicker.Stop()
	defer sendTicker.Stop()
}
