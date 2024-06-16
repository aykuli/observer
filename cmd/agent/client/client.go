package client

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/go-resty/resty/v2"

	"github.com/aykuli/observer/internal/agent/config"
	"github.com/aykuli/observer/internal/models"
	"github.com/aykuli/observer/internal/sign"

	"github.com/aykuli/observer/internal/agent/storage"
	"github.com/aykuli/observer/internal/compressor"
)

type MetricsClient struct {
	ServerAddr string
	memStorage *storage.MemStorage
	signKey    string
	limit      int
}

func NewMetricsClient(config config.Config, memStorage *storage.MemStorage) *MetricsClient {
	return &MetricsClient{
		ServerAddr: "http://" + config.Address,
		memStorage: memStorage,
		signKey:    config.Key,
		limit:      config.RateLimit,
	}
}

const (
	RetryCount              = 3
	RetryMinWaitTimeSeconds = 1
	RetryMaxWaitTimeSeconds = 5
)

func newRestyClient() *resty.Client {
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
	return restyClient
}

func (m *MetricsClient) SendMetrics() {
	metrics := m.memStorage.GetAllMetrics()

	if m.limit <= 0 {
		for i, mt := range metrics {
			if err := m.sendOneMetric(mt, i, nil); err != nil {
				log.Printf("Err sending metric %s with err %+v", mt.ID, err)
				return
			}
		}
		m.memStorage.ResetCounter()
		return
	}

	var wg sync.WaitGroup
	jobs := make(chan models.Metric, m.limit)
	errs := make(chan error, 1)
	fmt.Println("SIZE: ", len(metrics))
	for i, metric := range metrics {
		fmt.Println(i, " --- ", metric)
		jobs <- metric
		wg.Add(1)

		mt := <-jobs
		if err := m.sendOneMetric(mt, i, &wg); err != nil {
			log.Printf("Err sending metric %s with err %+v", mt.ID, err)
			errs <- err
			wg.Done()
		}

		fmt.Printf("wg: %+v\n", &wg)
	}
	fmt.Println("all metrics in jobs")

	fmt.Println("close(jobs)")
	close(jobs)
	fmt.Println("wg wait here")
	wg.Wait()

	if err := <-errs; err != nil {
		fmt.Println("err happened")
		// if any error happened we won't reset counter
		return
	}
	// no err in all goroutines - we can reset counter
	m.memStorage.ResetCounter()
}

func (m *MetricsClient) sendOneMetric(metric models.Metric, k int, wg *sync.WaitGroup) error {
	defer wg.Done()

	fmt.Println(k, "sendOneMetric ", metric.ID)
	req := newRestyClient().R()

	req.SetHeader("Content-Type", "application/json")
	req.URL = fmt.Sprintf("%s/update/", m.ServerAddr)
	req.Method = http.MethodPost

	marshalled, err := json.Marshal(metric)
	if err != nil {
		return err
	}

	if m.signKey != "" {
		req.SetHeader("HashSHA256", sign.GetHmacString(marshalled, m.signKey))
	}

	gzipped, err := compressor.Compress(marshalled)
	if err != nil {
		return err
	}

	if _, err = req.SetBody(gzipped).Send(); err != nil {
		return err
	}
	fmt.Println(k, "finish sendOneMetric ", metric.ID, "\n")

	return nil
}

func (m *MetricsClient) SendBatchMetrics(req *resty.Request) {
	req.SetHeader("Content-Type", "application/json")
	req.URL = fmt.Sprintf("%s/updates/", m.ServerAddr)
	req.Method = http.MethodPost

	metrics := m.memStorage.GetAllMetrics()
	if len(metrics) == 0 {
		return
	}

	marshalled, err := json.Marshal(metrics)
	if err != nil {
		log.Printf("Err JSON marshalling metrics with err %+v", err)
		return
	}

	if m.signKey != "" {
		req.SetHeader("HashSHA256", sign.GetHmacString(marshalled, m.signKey))
	}

	gzipped, err := compressor.Compress(marshalled)
	if err != nil {
		log.Printf("Err compressing metrics with err %+v", err)
		return
	}

	if _, err := req.SetBody(gzipped).Send(); err != nil {
		log.Printf("Err sending metrics with err %+v", err)
	} else {
		m.memStorage.ResetCounter()
	}
}
