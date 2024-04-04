package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/aykuli/observer/internal/server/config"
	"github.com/aykuli/observer/internal/server/db/postgres"
	"github.com/aykuli/observer/internal/server/logger"
	"github.com/aykuli/observer/internal/server/models"
	"github.com/aykuli/observer/internal/server/storage"
)

type Metrics struct {
	MemStorage *storage.MemStorage
}

func (m *Metrics) GetAllMetrics() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		var metrics []string
		for ck := range m.MemStorage.CounterMetrics {
			metrics = append(metrics, ck)
		}
		for gk := range m.MemStorage.GaugeMetrics {
			metrics = append(metrics, gk)
		}

		rw.Header().Set("Content-Type", "text/html")
		_, err := rw.Write([]byte(strings.Join(metrics, " ")))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
		}
	}
}

func (m *Metrics) GetMetric() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")

		var resultValue string
		switch metricType {
		case "gauge":
			value, ok := m.MemStorage.GaugeMetrics[metricName]
			if ok {
				resultValue = fmt.Sprintf("%v", value)
			} else {
				http.Error(w, "No such metric", http.StatusNotFound)
			}

		case "counter":
			value, ok := m.MemStorage.CounterMetrics[metricName]
			if ok {
				resultValue = fmt.Sprintf("%v", value)
			} else {
				http.Error(w, "No such metric", http.StatusNotFound)
			}
		default:
			http.Error(w, "No such metric", http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(resultValue))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
}

func checkType(metricType string) bool {
	if metricType == "gauge" || metricType == "counter" {
		return true
	}

	return false
}

func (m *Metrics) Update() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(rw, "Only POST method allowed", http.StatusMethodNotAllowed)
			return
		}

		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")
		metricValue := chi.URLParam(r, "metricValue")

		if !checkType(metricType) {
			http.Error(rw, "Metric type is wrong", http.StatusBadRequest)
			return
		}

		if metricName == "" {
			http.Error(rw, "Metric name is empty", http.StatusNotFound)
			return
		}

		switch metricType {
		case "gauge":
			value, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				http.Error(rw, "Metric value is wrong", http.StatusBadRequest)
				return
			}

			m.MemStorage.GaugeMetrics[metricName] = value
		case "counter":
			value, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				http.Error(rw, "Metric value is wrong", http.StatusBadRequest)
				return
			}

			m.MemStorage.CounterMetrics[metricName] += value
		default:
			http.Error(rw, "No such metric type", http.StatusNotFound)
			return

		}

		rw.Header().Set("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusOK)
	}
}

func (m *Metrics) UpdateFromJSON() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		var metric models.Metrics
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&metric); err != nil {
			logger.Log.Debug("cannot decode json request body", zap.Error(err))
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		metricName := metric.ID
		metricType := metric.MType

		if !checkType(metricType) {
			http.Error(rw, "Metric type is wrong", http.StatusBadRequest)
			return
		}

		if metricName == "" {
			http.Error(rw, "Metric name is empty", http.StatusNotFound)
			return
		}
		var gaugeValue *float64
		var countValue int64
		switch metricType {
		case "gauge":
			m.MemStorage.GaugeMetrics[metricName] = *metric.Value
			gaugeValue = metric.Value
		case "counter":
			m.MemStorage.CounterMetrics[metricName] += *metric.Delta
			countValue = m.MemStorage.CounterMetrics[metricName]
		default:
			http.Error(rw, "No such metric type", http.StatusNotFound)
			return
		}

		if config.Options.SaveMetrics && config.Options.StoreInterval == 0 {
			if err := m.MemStorage.SaveMetricsToFile(); err != nil {
				log.Print(err)
			}
		}

		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		var respMetric models.Metrics
		switch metricType {
		case "gauge":
			respMetric = models.Metrics{
				ID:    metricName,
				MType: metricType,
				Value: gaugeValue,
			}
		case "counter":
			respMetric = models.Metrics{
				ID:    metricName,
				MType: metricType,
				Delta: &countValue,
			}
		}

		resp, err := json.Marshal(respMetric)
		if err != nil {
			logger.Log.Debug("cannot encode json request body", zap.Error(err))
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, err = rw.Write(resp)
		if err != nil {
			logger.Log.Debug("cannot encode json request body", zap.Error(err))
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func (m *Metrics) ReadMetric() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var askedMetric models.Metrics
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&askedMetric); err != nil {
			logger.Log.Debug("cannot decode json request body", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		metricName := askedMetric.ID
		metricType := askedMetric.MType

		var respMetric = models.Metrics{ID: askedMetric.ID, MType: askedMetric.MType}
		switch metricType {
		case "gauge":
			value := m.MemStorage.GaugeMetrics[metricName]
			respMetric.Value = &value
		case "counter":
			value := m.MemStorage.CounterMetrics[metricName]
			respMetric.Delta = &value
		default:
			http.Error(w, "No such metric", http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
		enc := json.NewEncoder(w)
		if err := enc.Encode(respMetric); err != nil {
			logger.Log.Debug("cannot encode json request body", zap.Error(err))
			w.WriteHeader(http.StatusUnprocessableEntity)
			return

		}

		defer r.Body.Close()
	}
}

func (m *Metrics) Ping() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := postgres.Instance.Ping(r.Context())
		if err != nil {
			logger.Log.Debug("database connection invalid", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte("pong"))
		if err != nil {
			logger.Log.Debug("something went wrong", zap.Error(err))
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		defer r.Body.Close()
	}
}
