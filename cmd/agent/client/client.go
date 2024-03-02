package client

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-resty/resty/v2"

	"github.com/aykuli/observer/internal/models"
	"github.com/aykuli/observer/internal/sign"

	"github.com/aykuli/observer/internal/agent/storage"
	"github.com/aykuli/observer/internal/compressor"
)

type MerticsClient struct {
	ServerAddr string
	memStorage *storage.MemStorage
	signKey    string
	limit      int
}

func NewMetricsClint(serverAddr string, memStorage *storage.MemStorage) *MerticsClient {
	return &MerticsClient{
		ServerAddr: serverAddr,
		memStorage: memStorage,
	}
}

// SendMetrics deprecated from iteration 5
func (m *MerticsClient) SendMetrics(req *resty.Request) {
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

	jobs := make(chan models.Metric, m.limit)
	for i, metric := range metrics {
		jobs <- metric
		if i == len(metrics)-1 {
			m.memStorage.ResetCounter()
		}
	}

	for w := 0; w < m.limit; w++ {
		go func() {
			for metric := range jobs {
				if err := m.sendOneMetric(req, metric); err != nil {
					log.Printf("Err sending metric %s with err %+v", metric.ID, err)
					return
				}
			}
		}()
	}

	close(jobs)
}

func (m *MerticsClient) sendOneMetric(req *resty.Request, metric models.Metric) error {
	req.SetHeader("Content-Type", "application/json")
	req.URL = fmt.Sprintf("%s/update/", m.ServerAddr)
	req.Method = http.MethodPost

	marshalled, err := json.Marshal(metric)
	if err != nil {
		return err
	}

	req.SetHeader("HashSHA256", sign.GetHmacString(marshalled, m.signKey))

	gzipped, err := compressor.Compress(marshalled)
	if err != nil {
		return err
	}

	if _, err = req.SetBody(gzipped).Send(); err != nil {
		return err
	}

	return nil
}

func (m *MerticsClient) SendBatchMetrics(req *resty.Request) {
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

	req.SetHeader("HashSHA256", sign.GetHmacString(marshalled, m.signKey))

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
