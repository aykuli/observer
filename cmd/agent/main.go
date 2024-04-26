package main

import (
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/aykuli/observer/cmd/agent/client"
	"github.com/aykuli/observer/internal/agent/config"
	"github.com/aykuli/observer/internal/agent/storage"
)

const (
	RetryCount              = 3
	RetryMinWaitTimeSeconds = 1
	RetryMaxWaitTimeSeconds = 5
)

func main() {
	restyClient := resty.New().
		SetRetryCount(RetryCount).
		SetRetryWaitTime(RetryMinWaitTimeSeconds).
		SetRetryMaxWaitTime(RetryMaxWaitTimeSeconds).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			isConnRefused := r.StatusCode() == 0
			isServerDBErr := r.StatusCode() == http.StatusInternalServerError
			return isConnRefused || isServerDBErr
		})
	restyClient.OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
		r.SetHeader("Content-Encoding", "gzip")
		r.SetHeader("Accept-Encoding", "gzip")

		return nil
	})
	request := restyClient.R()

	memStorage := storage.NewMemStorage()
	newClient := client.NewMetricsClint("http://"+config.Options.Address, memStorage)

	collectTicker := time.NewTicker(time.Duration(config.Options.PollInterval) * time.Second)
	collectQuit := make(chan struct{})
	sendTicker := time.NewTicker(time.Duration(config.Options.ReportInterval) * time.Second)
	sendQuit := make(chan struct{})
	defer collectTicker.Stop()
	defer sendTicker.Stop()

	for {
		select {
		case <-collectTicker.C:
			memStorage.GarbageStats()
		case <-sendTicker.C:
			newClient.SendMetrics(request)
			newClient.SendBatchMetrics(request)
		case <-sendQuit:
			sendTicker.Stop()
			return
		case <-collectQuit:
			collectTicker.Stop()
			return
		}
	}
}
