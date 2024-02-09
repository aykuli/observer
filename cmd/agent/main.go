package main

import (
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/aykuli/observer/cmd/agent/handlers"
	"github.com/aykuli/observer/internal/storage"
)

func main() {
	parseFlags()
	request := resty.New().R()

	memStorage := storage.MemStorage{GaugeMetrics: map[string]float64{}, CounterMetrics: map[string]int64{}}
	collectTicker := time.NewTicker(pollInterval)
	collectQuit := make(chan struct{})

	sendTicker := time.NewTicker(reportInterval)
	sendQuit := make(chan struct{})

	i := 0
	for i < 5 {
		i++
		for {
			select {
			case <-collectTicker.C:
				storage.GetStats(&memStorage)
			case <-sendTicker.C:
				handlers.SendPost(request, addr, memStorage)
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
