package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/aykuli/observer/internal/storage"
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

		rw.Header().Set("Content-Type", "text/plain")
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
