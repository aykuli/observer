package main

import (
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/aykuli/observer/cmd/agent/client"
	"github.com/aykuli/observer/internal/agent/config"
	"github.com/aykuli/observer/internal/agent/storage"
)

func main() {
	restyClient := resty.New()
	restyClient.OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
		r.SetHeader("Content-Encoding", "gzip")
		r.SetHeader("Accept-Encoding", "gzip")

		return nil
	})
	request := restyClient.R()

	memStorage := storage.NewMemStorage()
	collectTicker := time.NewTicker(time.Duration(config.Options.PollInterval) * time.Second)
	collectQuit := make(chan struct{})

	sendTicker := time.NewTicker(time.Duration(config.Options.ReportInterval) * time.Second)
	sendQuit := make(chan struct{})

	newClient := client.MerticsClient{
		ServerAddr: "http://" + config.Options.Address,
		MemStorage: memStorage,
	}

	fmt.Printf("agent config: %+v\n\n", config.Options)

	defer collectTicker.Stop()
	defer sendTicker.Stop()

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
