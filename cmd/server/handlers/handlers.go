package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/aykuli/observer/internal/errors"
	"github.com/aykuli/observer/internal/server/config"
	"github.com/aykuli/observer/internal/server/db/postgres"
	"github.com/aykuli/observer/internal/server/logger"
	"github.com/aykuli/observer/internal/server/models"
	"github.com/aykuli/observer/internal/server/repository"
	"github.com/aykuli/observer/internal/server/storage"
	"github.com/aykuli/observer/internal/sign"
)

type Metric struct {
	MemStorage *storage.MemStorage
}

func (m *Metric) GetAllMetrics() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		var metricNames []string

		if config.Options.DatabaseDsn == "" {
			for ck := range m.MemStorage.CounterMetrics {
				metricNames = append(metricNames, ck)
			}
			for gk := range m.MemStorage.GaugeMetrics {
				metricNames = append(metricNames, gk)
			}
		} else {
			metricNamesFromDB, err := m.getMetricNamesFromDB(r.Context())
			if err != nil {
				http.Error(rw, errors.NewDBError(err).Error(), http.StatusInternalServerError)
				return
			}

			metricNames = metricNamesFromDB
		}

		rw.Header().Set("Content-Type", "text/html")
		rw.WriteHeader(http.StatusOK)
		_, err := rw.Write([]byte(strings.Join(metricNames, " ")))
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
			http.Error(w, errors.NewDBError(err).Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte("pong"))
		if err != nil {
			logger.Log.Debug("response body writing error", zap.Error(err))
			http.Error(w, errors.NewDBError(err).Error(), http.StatusUnprocessableEntity)
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
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
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
			metric, err := m.getMetricFromDB(r.Context(), metricName)
			if err != nil {
				http.Error(w, errors.NewDBError(err).Error(), http.StatusNotFound)
				return
			}

			respMetric = metric
		}

		w.WriteHeader(http.StatusOK)
		enc := json.NewEncoder(w)
		if err := enc.Encode(respMetric); err != nil {
			logger.Log.Debug("cannot encode json request body", zap.Error(err))
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
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
					return
				}
			case "counter":
				delta, ok := m.MemStorage.CounterMetrics[metricName]
				if ok {
					resultValue = fmt.Sprintf("%v", delta)
				} else {
					http.Error(w, "No such metric", http.StatusNotFound)
					return
				}
			default:
				http.Error(w, "No such metric", http.StatusNotFound)
				return
			}
		} else {
			metric, err := m.getMetricFromDB(r.Context(), metricName)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
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
	return func(w http.ResponseWriter, r *http.Request) {
		var metric models.Metric
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&metric); err != nil {
			logger.Log.Debug("cannot decode json request body", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if config.Options.DatabaseDsn != "" {
			err := m.insertIntoDB(r.Context(), metric)
			if err != nil {
				logger.Log.Debug("cannot save to database", zap.Error(err))
				http.Error(w, errors.NewDBError(err).Error(), http.StatusInternalServerError)
				return
			}
		} else {
			switch metric.MType {
			case "gauge":
				m.MemStorage.GaugeMetrics[metric.ID] = *metric.Value
			case "counter":
				m.MemStorage.CounterMetrics[metric.ID] += *metric.Delta
			default:
				http.Error(w, "no such metric type", http.StatusInternalServerError)
				return
			}

			if config.Options.FileStoragePath != "" && config.Options.StoreInterval == 0 {
				err := m.MemStorage.SaveMetricsToFile()
				if err != nil {
					logger.Log.Debug("localhost save err", zap.Error(err))
					http.Error(w, "metrics wasnt saved", http.StatusInternalServerError)
					return
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Encoding", "gzip")
		w.WriteHeader(http.StatusOK)

		resp, err := json.Marshal(metric)
		if err != nil {
			logger.Log.Debug("cannot encode json request body", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = w.Write(resp)
		if err != nil {
			logger.Log.Debug("cannot encode json request body", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (m *Metric) Update() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method allowed", http.StatusMethodNotAllowed)
			return
		}

		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")
		metricValue := chi.URLParam(r, "metricValue")

		if !checkType(metricType) {
			http.Error(w, "Metric type is wrong", http.StatusBadRequest)
			return
		}

		if metricName == "" {
			http.Error(w, "Metric name is empty", http.StatusNotFound)
			return
		}

		switch metricType {
		case "gauge":
			value, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				http.Error(w, "Metric value is wrong", http.StatusBadRequest)
				return
			}

			m.MemStorage.GaugeMetrics[metricName] = value
		case "counter":
			value, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				http.Error(w, "Metric value is wrong", http.StatusBadRequest)
				return
			}

			m.MemStorage.CounterMetrics[metricName] += value
		default:
			http.Error(w, "No such metric type", http.StatusNotFound)
			return
		}

		if config.Options.FileStoragePath != "" && config.Options.StoreInterval == 0 {
			err := m.MemStorage.SaveMetricsToFile()
			if err != nil {
				logger.Log.Debug("cannot save to database", zap.Error(err))
				http.Error(w, errors.NewDBError(err).Error(), http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
	}
}

func (m *Metric) BatchUpdate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var metrics []models.Metric
		if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
			logger.Log.Debug("cannot decode json request body", zap.Error(err))
			http.Error(w, errors.NewDBError(err).Error(), http.StatusInternalServerError)
			return
		}

		if config.Options.Key != "" {
			byteData, err := json.Marshal(metrics)
			if err != nil {
				logger.Log.Debug("cannot marshal metrics", zap.Error(err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			signString := r.Header.Get("HashSHA256")
			equal, err := sign.Verify(byteData, config.Options.Key, signString)
			if err != nil {
				logger.Log.Debug("sign verifying error", zap.Error(err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if !equal {
				logger.Log.Debug("signs not equal", zap.Error(err))
				http.Error(w, "cannot serve this agent", http.StatusBadRequest)
				return
			}
		}

		if config.Options.DatabaseDsn != "" {
			err := m.insertManyIntoDB(r.Context(), metrics)
			if err != nil {
				logger.Log.Debug("cannot save to database", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			for _, mt := range metrics {
				switch mt.MType {
				case "gauge":
					m.MemStorage.GaugeMetrics[mt.ID] = *mt.Value
				case "counter":
					m.MemStorage.CounterMetrics[mt.ID] += *mt.Delta
				default:
					http.Error(w, "No such metric type", http.StatusNotFound)
					return

				}
			}

			if config.Options.FileStoragePath != "" && config.Options.StoreInterval == 0 {
				err := m.MemStorage.SaveMetricsToFile()
				if err != nil {
					logger.Log.Debug("cannot save to database", zap.Error(err))
					http.Error(w, errors.NewDBError(err).Error(), http.StatusInternalServerError)
					return
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Encoding", "gzip")
		w.WriteHeader(http.StatusOK)

		resp, err := json.Marshal(metrics)
		if err != nil {
			logger.Log.Debug("cannot encode json request body", zap.Error(err))
			http.Error(w, errors.NewDBError(err).Error(), http.StatusInternalServerError)
			return
		}

		if config.Options.Key != "" {
			signString, err := sign.GetHmacString(resp, config.Options.Key)
			if err != nil {
				log.Printf("Err singing body with err %+v", err)
				return
			}
			r.Header.Set("HashSHA256", signString)
		}

		_, err = w.Write(resp)
		if err != nil {
			logger.Log.Debug("cannot write response", zap.Error(err))
			http.Error(w, errors.NewDBError(err).Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (m *Metric) NotFound() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusNotFound)
	}
}

func (m *Metric) BadRequest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusBadRequest)
	}
}

func checkType(metricType string) bool {
	if metricType == "gauge" || metricType == "counter" {
		return true
	}

	return false
}

func (m *Metric) getMetricNamesFromDB(ctx context.Context) ([]string, error) {
	conn, err := postgres.Instance.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	var metrics []string
	metricsNamesRepo := repository.NewMetricNamesRepository(conn)
	metricNames, err := metricsNamesRepo.SelectAll(ctx)
	if err != nil {
		return nil, err
	}
	for _, m := range metricNames {
		metrics = append(metrics, m.Name)
	}

	return metrics, nil
}

func (m *Metric) insertIntoDB(ctx context.Context, metric models.Metric) error {
	conn, err := postgres.Instance.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	metricNamesRepo := repository.NewMetricNamesRepository(conn)
	metricName, err := metricNamesRepo.FindByName(ctx, metric.ID)
	if err != nil {
		return err
	}

	metricsRepo := repository.NewMetricsRepository(conn)
	err = metricsRepo.Insert(ctx, metricName.ID, metric)
	if err != nil {
		return err
	}

	return nil
}

func (m *Metric) insertManyIntoDB(ctx context.Context, metrics []models.Metric) error {
	conn, err := postgres.Instance.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	var names []string
	for _, mt := range metrics {
		names = append(names, mt.ID)
	}

	metricNamesRepo := repository.NewMetricNamesRepository(conn)
	metricNames, err := metricNamesRepo.SelectByNames(ctx, names)
	if err != nil {
		return err
	}

	metricNamesMap := make(map[string]int)
	for _, mn := range metricNames {
		metricNamesMap[mn.Name] = mn.ID
	}

	metricsRepo := repository.NewMetricsRepository(conn)
	err = metricsRepo.InsertBatch(ctx, metrics, metricNamesMap)
	if err != nil {
		return err
	}

	return nil
}

func (m *Metric) getMetricFromDB(ctx context.Context, metricName string) (*models.Metric, error) {
	conn, err := postgres.Instance.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	metricNamesRepo := repository.NewMetricNamesRepository(conn)
	metricNameDB, err := metricNamesRepo.FindByName(ctx, metricName)
	if err != nil {
		return nil, err
	}

	metricsRepo := repository.NewMetricsRepository(conn)
	metric, err := metricsRepo.FindByMetricName(ctx, metricNameDB)
	if err != nil {
		return nil, err
	}
	return metric, nil
}
