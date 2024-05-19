package client

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-resty/resty/v2"

	"github.com/aykuli/observer/internal/models"

	"github.com/aykuli/observer/internal/agent/storage"
	"github.com/aykuli/observer/internal/compressor"
)

type MerticsClient struct {
	ServerAddr string
	memStorage *storage.MemStorage
}

func NewMetricsClint(serverAddr string, memStorage *storage.MemStorage) *MerticsClient {
	return &MerticsClient{
		ServerAddr: serverAddr,
		memStorage: memStorage,
	}
}

// SendMetrics deprecated from iteration 5
func (m *MerticsClient) SendMetrics(req *resty.Request) {
	for _, mt := range m.memStorage.GetAllMetrics() {
		err := m.sendOneMetric(req, mt)
		if err != nil {
			log.Printf("Err sending gauge metric %s with err %+v", mt.ID, err)
			return
		}
	}
	m.memStorage.ResetCounter()
}

func (m *MerticsClient) sendOneMetric(req *resty.Request, metric models.Metric) error {
	req.SetHeader("Content-Type", "application/json")
	req.URL = fmt.Sprintf("%s/update/", m.ServerAddr)
	req.Method = http.MethodPost

	gzipped, err := compressor.Compress(metric)
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

	if len(metrics) > 0 {
		gzipped, err := compressor.Compress(metrics)
		if err != nil {
			log.Printf("Err compressing metrics with err %+v", err)
			return
		}

		if _, err = req.SetBody(gzipped).Send(); err != nil {
			log.Printf("Err sending metrics with err %+v", err)
		} else {
			m.memStorage.ResetCounter()
		}
	}
}
