package client

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-resty/resty/v2"

	"github.com/aykuli/observer/internal/agent/config"
	"github.com/aykuli/observer/internal/agent/models"
	"github.com/aykuli/observer/internal/agent/storage"
	"github.com/aykuli/observer/internal/compressor"
	"github.com/aykuli/observer/internal/sign"
)

var (
	gaugeMType   = "gauge"
	counterMType = "counter"
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
		body := models.Metric{
			ID:    k,
			MType: gaugeMType,
			Delta: nil,
			Value: &v,
		}
		marshalled, err := json.Marshal(body)
		if err != nil {
			log.Printf("Err JSON marshalling metrics with err %+v", err)
			return
		}

		gzipped, err := compressor.Compress(marshalled)
		if err != nil {
			log.Printf("Err gzipping gauge metric %s with err %+v", k, err)
			return
		}

		_, err = req.SetBody(gzipped).Send()
		if err != nil {
			log.Printf("Err sending gauge metric %s with err %+v", k, err)
		}
	}

	for k, v := range m.MemStorage.CounterMetrics {
		body := models.Metric{
			ID:    k,
			MType: counterMType,
			Delta: &v,
			Value: nil,
		}
		marshalled, err := json.Marshal(body)
		if err != nil {
			log.Printf("Err JSON marshalling metrics with err %+v", err)
			return
		}
		gzipped, err := compressor.Compress(marshalled)
		if err != nil {
			log.Printf("Err sending counter metric %s with err %+v", k, err)
			return
		}

		if _, err = req.SetBody(gzipped).Send(); err != nil {
			log.Printf("Err sending counter metric %s with err %+v", k, err)
		}
	}
}

func (m *MerticsClient) SendBatchMetrics(req *resty.Request) {
	req.SetHeader("Content-Type", "application/json")
	req.URL = fmt.Sprintf("%s/updates/", m.ServerAddr)
	req.Method = http.MethodPost

	var metrics []models.Metric
	for k := range m.MemStorage.GaugeMetrics {
		v := m.MemStorage.GaugeMetrics[k]
		metrics = append(metrics, models.Metric{
			ID:    k,
			MType: gaugeMType,
			Delta: nil,
			Value: &v,
		})
	}

	for k := range m.MemStorage.CounterMetrics {
		d := m.MemStorage.CounterMetrics[k]
		metrics = append(metrics, models.Metric{
			ID:    k,
			MType: counterMType,
			Delta: &d,
			Value: nil,
		})
	}

	if len(metrics) > 0 {
		marshalled, err := json.Marshal(metrics)
		if err != nil {
			log.Printf("Err JSON marshalling metrics with err %+v", err)
			return
		}

		if config.Options.Key != "" {
			signString, err := sign.GetHmacString(marshalled, config.Options.Key)
			if err != nil {
				log.Printf("Err singing body with err %+v", err)
				return
			}
			req.Header.Set("HashSHA256", signString)
		}

		if config.Options.Key != "" {
			h := hmac.New(sha256.New, []byte(config.Options.Key))
			_, err = h.Write(marshalled)
			if err == nil {
				signString := hex.EncodeToString(h.Sum(nil))
				req.Header.Set("HashSHA256", signString)
			} else {
				log.Printf("Cannot sign marshalled body with err %+v", err)
			}
		}

		gzipped, err := compressor.Compress(marshalled)
		if err != nil {
			log.Printf("Err compressing metrics with err %+v", err)
			return
		}

		if _, err := req.SetBody(gzipped).Send(); err != nil {
			log.Printf("Err sending metrics with err %+v", err)
		}
	}
}

func getSignString(body []byte, key string) (string, error) {
	h := hmac.New(sha256.New, []byte(key))
	_, err := h.Write(body)
	if err == nil {
		return hex.EncodeToString(h.Sum(nil)), nil
	} else {
		return "", err
	}
}
