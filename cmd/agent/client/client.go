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
	MemStorage storage.MemStorage
}

func (m *MerticsClient) SendMetrics(req *resty.Request) {
	req.SetHeader("Content-Type", "application/json")
	req.URL = fmt.Sprintf("%s/update/", m.ServerAddr)
	req.Method = http.MethodPost

	for k, v := range m.MemStorage.GaugeMetrics {
		m.sendOneMetric(req, models.Metric{ID: k, MType: "gauge", Delta: nil, Value: &v})
	}

	for k, v := range m.MemStorage.CounterMetrics {
		m.sendOneMetric(req, models.Metric{ID: k, MType: "counter", Delta: &v, Value: nil})
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

	var metrics []models.Metric
	for k := range m.MemStorage.GaugeMetrics {
		v := m.MemStorage.GaugeMetrics[k]
		metrics = append(metrics, models.Metric{ID: k, MType: "gauge", Delta: nil, Value: &v})
	}

	for k := range m.MemStorage.CounterMetrics {
		d := m.MemStorage.CounterMetrics[k]
		metrics = append(metrics, models.Metric{ID: k, MType: "counter", Delta: &d, Value: nil})
	}

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
