package handlers

import (
	"context"
	"encoding/json"
	"errors"
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
	"github.com/aykuli/observer/internal/server/repository/metric_names_repository"
	"github.com/aykuli/observer/internal/server/repository/metrics_repository"
	"github.com/aykuli/observer/internal/server/storage"
)

type Metric struct {
	MemStorage *storage.MemStorage
}

func (m *Metric) GetAllMetrics() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		var metrics []string

		if config.Options.DatabaseDsn == "" {
			for ck := range m.MemStorage.CounterMetrics {
				metrics = append(metrics, ck)
			}
			for gk := range m.MemStorage.GaugeMetrics {
				metrics = append(metrics, gk)
			}
		} else {
			metricNames, err := m.getMetricNamesFromDB(r.Context())
			if err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}

			metrics = metricNames
		}

		rw.Header().Set("Content-Type", "text/html")
		_, err := rw.Write([]byte(strings.Join(metrics, " ")))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
		}
	}
}

func (m *Metric) Ping() http.HandlerFunc {
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

func (m *Metric) ReadMetric() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var askedMetric models.Metric
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&askedMetric); err != nil {
			logger.Log.Debug("cannot decode json request body", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		metricName := askedMetric.ID
		metricType := askedMetric.MType

		var respMetric = &models.Metric{ID: askedMetric.ID, MType: askedMetric.MType}
		if config.Options.DatabaseDsn == "" {
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
		} else {
			metric, err := m.getMetricFromDB(r.Context(), metricName, metricType)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}

			respMetric = metric
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

func (m *Metric) GetMetric() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricName := chi.URLParam(r, "metricName")
		metricType := chi.URLParam(r, "metricType")
		var resultValue string

		if config.Options.DatabaseDsn == "" {
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
		} else {
			metric, err := m.getMetricFromDB(r.Context(), metricName, metricType)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
			if metric.Value != nil {
				resultValue = fmt.Sprintf("%v", metric.Value)
			} else {
				resultValue = fmt.Sprintf("%v", metric.Delta)
			}

		}

		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(resultValue))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
}

func (m *Metric) UpdateFromJSON() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		var metric models.Metric
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&metric); err != nil {
			logger.Log.Debug("cannot decode json request body", zap.Error(err))
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		//metricName := metric.ID
		//
		//if metricName == "" {
		//	http.Error(rw, "Metric name is empty", http.StatusNotFound)
		//	return
		//}

		if config.Options.DatabaseDsn != "" {
			err := m.saveIntoDB(r.Context(), metric)
			if err != nil {
				logger.Log.Debug("cannot save to database", zap.Error(err))
				rw.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			err := m.saveOnLocalhost(metric)
			if err != nil {
				logger.Log.Debug("localhost save err", zap.Error(err))
				log.Print(err)
			}
		}

		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)

		resp, err := json.Marshal(metric)
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

func (m *Metric) Update() http.HandlerFunc {
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

func (m *Metric) NotFound() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusNotFound)
	}
}

func (m *Metric) BadRequest() http.HandlerFunc {
	return func(rw http.ResponseWriter, w *http.Request) {
		rw.Header().Set("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusBadRequest)
	}
}

func checkType(metricType string) bool {
	if metricType == "gauge" || metricType == "counter" {
		return true
	}

	return false
}

// helper func for m.GetAllMetrics
func (m *Metric) getMetricNamesFromDB(ctx context.Context) ([]string, error) {
	conn, err := postgres.Instance.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	var metrics []string
	metricsNamesRepo := metric_names_repository.NewRepository(conn)
	metricNames, err := metricsNamesRepo.SelectAll(ctx)
	if err != nil {
		return nil, err
	}
	for _, m := range metricNames {
		metrics = append(metrics, m.Name)
	}

	return metrics, nil
}

func (m *Metric) saveIntoDB(ctx context.Context, metric models.Metric) error {
	conn, err := postgres.Instance.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	metricNamesRepo := metric_names_repository.NewRepository(conn)
	metricID, err := metricNamesRepo.GetID(ctx, metric.ID)
	if err != nil {
		return err
	}

	metricsRepo := metrics_repository.NewRepository(conn)
	err = metricsRepo.Insert(ctx, metricID, metric)
	if err != nil {
		return err
	}

	return nil
}

func (m *Metric) saveOnLocalhost(metric models.Metric) error {
	switch metric.MType {
	case "gauge":
		m.MemStorage.GaugeMetrics[metric.ID] = *metric.Value
	case "counter":
		m.MemStorage.CounterMetrics[metric.ID] += *metric.Delta
	default:
		return errors.New("no such metric type")
	}

	if config.Options.SaveMetrics && config.Options.StoreInterval == 0 {
		if err := m.MemStorage.SaveMetricsToFile(); err != nil {
			return err
		}
	}

	return nil
}

func (m *Metric) getMetricFromDB(ctx context.Context, metricName string, metricType string) (*models.Metric, error) {
	conn, err := postgres.Instance.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	metricNamesRepo := metric_names_repository.NewRepository(conn)
	metricID, err := metricNamesRepo.GetID(ctx, metricName)
	if err != nil {
		return nil, err
	}

	metricsRepo := metrics_repository.NewRepository(conn)
	metric, err := metricsRepo.FindByMetricId(ctx, metricID)
	if err != nil {
		return nil, err
	}
	return metric, nil
}
