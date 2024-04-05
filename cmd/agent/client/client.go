package client

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-resty/resty/v2"

	"github.com/aykuli/observer/internal/agent/models"
	"github.com/aykuli/observer/internal/agent/storage"
)

type MerticsClient struct {
	ServerAddr string
	MemStorage storage.MemStorage
}

func (m *MerticsClient) SendBatchMetrics(req *resty.Request) {
	req.SetHeader("Content-Type", "application/json")
	req.URL = fmt.Sprintf("%s/updates/", m.ServerAddr)
	req.Method = http.MethodPost

	var metrics []*models.Metric
	for k, v := range m.MemStorage.GaugeMetrics {
		metrics = append(metrics, makeMetricValue(k, "gauge", v))
	}

	for k, d := range m.MemStorage.CounterMetrics {
		metrics = append(metrics, makeMetricDelta(k, "counter", d))
	}

	if len(metrics) > 0 {
		_, err := req.SetBody(metrics).Send()
		if err != nil {
			log.Printf("Err sending metrics %+v with err %+v", metrics, err)
		}
	}
}

func makeMetricValue(name string, mType string, value float64) *models.Metric {
	return &models.Metric{
		ID:    name,
		MType: mType,
		Delta: nil,
		Value: &value,
	}
}

func makeMetricDelta(name string, mType string, delta int64) *models.Metric {
	return &models.Metric{
		ID:    name,
		MType: mType,
		Delta: &delta,
		Value: nil,
	}
}

func (m *MerticsClient) SendMetrics(req *resty.Request) {
	req.SetHeader("Content-Type", "application/json")
	req.URL = fmt.Sprintf("%s/update/", m.ServerAddr)
	req.Method = http.MethodPost

	for k, v := range m.MemStorage.GaugeMetrics {
		body := models.Metric{
			ID:    k,
			MType: "gauge",
			Delta: nil,
			Value: &v,
		}

		_, err := req.SetBody(body).Send()
		if err != nil {
			log.Printf("Err sending gauge metric %s with err %+v", k, err)
		}
	}

	for k, v := range m.MemStorage.CounterMetrics {
		body := models.Metric{
			ID:    k,
			MType: "counter",
			Delta: &v,
			Value: nil,
		}

		_, err := req.SetBody(body).Send()
		if err != nil {
			log.Printf("Err sending counter metric %s with err %+v", k, err)
		}
	}
}

type Options struct {
	serverAddr, mType, mName, mValue string
}
