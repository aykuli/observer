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

func (m *MerticsClient) SendMetrics(req *resty.Request) {
	req.SetHeader("Content-Type", "application/json")
	req.URL = fmt.Sprintf("%s/update/", m.ServerAddr)
	req.Method = http.MethodPost

	for k, v := range m.MemStorage.GaugeMetrics {
		body := models.Metrics{
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
		body := models.Metrics{
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
