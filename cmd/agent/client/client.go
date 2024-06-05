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
func (m *MetricsClient) SendMetrics(req *resty.Request) {
	metrics := m.memStorage.GetAllMetrics()

	if m.limit <= 0 {
		for _, mt := range metrics {
			if err := m.sendOneMetric(req, mt); err != nil {
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

	for _, metric := range metrics {
		jobs <- metric
	}

	for w := 0; w < m.limit; w++ {
		wg.Add(1)
		go func() {
			for metric := range jobs {
				if err := m.sendOneMetric(req, metric); err != nil {
					errs <- err
					log.Printf("Err sending metric %s with err %+v", metric.ID, err)

					wg.Done()
				}
			}

			wg.Done()
		}()
	}

	if err := <-errs; err != nil {
		close(jobs)
		wg.Wait()
		return
	}

	close(jobs)
	m.memStorage.ResetCounter()
	wg.Wait()
}

func (m *MetricsClient) sendOneMetric(req *resty.Request, metric models.Metric) error {
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
	fmt.Printf("agent sendOneMetric req.Header.Get(\"HashSHA256\") %v\n", req.Header.Get("HashSHA256"))

	if _, err = req.SetBody(gzipped).Send(); err != nil {
		return err
	}

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
