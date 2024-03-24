package main

import (
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/aykuli/observer/cmd/agent/client"
	"github.com/aykuli/observer/internal/agent/config"
	"github.com/aykuli/observer/internal/storage"
)

func main() {
	restyClient := resty.New()
	restyClient.OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
		r.SetHeader("Content-Encoding", "gzip")
		r.SetHeader("Accept-Encoding", "gzip")

		return nil
	})
	request := restyClient.R()

	memStorage := storage.MemStorageInit
	collectTicker := time.NewTicker(time.Duration(config.PollInterval) * time.Second)
	collectQuit := make(chan struct{})

	sendTicker := time.NewTicker(time.Duration(config.ReportInterval) * time.Second)
	sendQuit := make(chan struct{})

	newClient := client.MerticsClient{
		ServerAddr: config.ListenAddr,
		MemStorage: memStorage,
	}

	i := 0
	for i < config.MaxTries {
		i++

		for {
			select {
			case <-collectTicker.C:
				storage.GetStats(&memStorage)
			case <-sendTicker.C:
				newClient.SendMetrics(request)
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
