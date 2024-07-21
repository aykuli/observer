package main

import (
	_ "net/http/pprof"
	"time"

	"github.com/aykuli/observer/cmd/agent/client"
	"github.com/aykuli/observer/internal/agent/config"
	"github.com/aykuli/observer/internal/agent/storage"
)

func main() {
	memStorage := storage.NewMemStorage()
	newClient := client.NewMetricsClient(config.Options, &memStorage)

	collectTicker := time.NewTicker(time.Duration(config.Options.PollInterval) * time.Second)
	sendTicker := time.NewTicker(time.Duration(config.Options.ReportInterval) * time.Second)
	defer collectTicker.Stop()
	defer sendTicker.Stop()

	for {
		select {
		case <-collectTicker.C:
			memStorage.GarbageStats()
			memStorage.GetSystemUtilInfo()
		case <-sendTicker.C:
			if config.Options.RateLimit > 0 {
				newClient.SendMetrics()
			} else {
				newClient.SendBatchMetrics()
			}
		}
	}
}
