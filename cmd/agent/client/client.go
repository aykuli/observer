package client

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-resty/resty/v2"

	"github.com/aykuli/observer/internal/agent/models"
	"github.com/aykuli/observer/internal/agent/storage"
	"github.com/aykuli/observer/internal/compressor"
)

type MerticsClient struct {
	ServerAddr string
	memStorage storage.MemStorage
}

func NewMetricsClint(serverAddr string, memStorage storage.MemStorage) *MerticsClient {
	return &MerticsClient{
		ServerAddr: serverAddr,
		memStorage: memStorage,
	}
}

func (m *MerticsClient) SendMetrics(req *resty.Request) {
	for _, mt := range m.memStorage.GetAllMetrics() {
		m.sendOneMetric(req, mt)
	}
}

func (m *MerticsClient) sendOneMetric(req *resty.Request, metric models.Metric) {
	req.SetHeader("Content-Type", "application/json")
	req.URL = fmt.Sprintf("%s/update/", m.ServerAddr)
	req.Method = http.MethodPost

	gzipped, err := compressor.Compress(metric)
	if err != nil {
		log.Printf("Err zipping metric %+v", err)
		return
	}

	_, err = req.SetBody(gzipped).Send()
	if err != nil {
		log.Printf("Err sending gauge metric %s with err %+v", metric.ID, err)
	}
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

		if _, err := req.SetBody(gzipped).Send(); err != nil {
			log.Printf("Err sending metrics with err %+v", err)
		}
	}
}
