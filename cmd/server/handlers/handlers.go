package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/aykuli/observer/internal/server/config"
	"github.com/aykuli/observer/internal/server/logger"
	"github.com/aykuli/observer/internal/server/models"
	"github.com/aykuli/observer/internal/server/storage"
	"github.com/aykuli/observer/internal/sign"
)

type APIV1 struct {
	Storage storage.Storage
}

// Ping godoc
//
//	@Produce		text/plain
//	@Success		200		{string}	json	"OK"
//	@Failure		422		{string}	error	"Unprocessable Entity"
//	@Failure		500		{string}	error	"Internal Server Error"
//	@Router			/ping [GET]
func (v *APIV1) Ping() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := v.Storage.Ping(r.Context())
		if err != nil {
			logger.Log.Debug("ping storage error", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte("pong"))
		if err != nil {
			logger.Log.Debug("response body writing error", zap.Error(err))
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		defer r.Body.Close()
	}
}

// GetAllMetrics godoc
//
//	@Produce		text/plain
//	@Success		200		{string}	json	"OK"
//	@Failure		400		{string}	error	"Bad Request"
//	@Failure		500		{string}	error	"Internal Server Error"
//	@Router			/ [GET]
func (v *APIV1) GetAllMetrics() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		metricsPage, err := v.Storage.GetMetrics(r.Context())
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		rw.Header().Set("Content-Type", "text/html")
		rw.WriteHeader(http.StatusOK)
		_, err = rw.Write([]byte(metricsPage))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
		}
	}
}

// ReadMetric godoc
//
//	@Accept			application/json
//	@Produce		application/json
//	@Success		200		{string}	json	"OK"
//	@Failure		400		{string}	error	"Bad Request"
//	@Failure		422		{string}	error	"Unprocessable Entity"
//	@Failure		500		{string}	error	"Internal Server Error"
//	@Router			/value [POST]
func (v *APIV1) ReadMetric() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var askedMetric models.Metric
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&askedMetric); err != nil {
			logger.Log.Debug("cannot decode json request body", zap.Error(err))
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		metric, err := v.Storage.ReadMetric(r.Context(), askedMetric.ID, askedMetric.MType)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		enc := json.NewEncoder(w)
		if err := enc.Encode(&metric); err != nil {
			logger.Log.Debug("cannot encode json request body", zap.Error(err))
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		}

		defer r.Body.Close()
	}
}

// GetMetric godoc
//
//	@Accept			application/json
//	@Produce		application/json
//	@Success		200		{string}	json	"OK"
//	@Failure		400		{string}	error	"Bad Request"
//	@Failure		404		{string}	error	"Not Found"
//	@Router			/value/{metricType}/{metricName} [GET]
func (v *APIV1) GetMetric() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mType := chi.URLParam(r, "metricType")
		mName := chi.URLParam(r, "metricName")
		metric, err := v.Storage.ReadMetric(r.Context(), mName, mType)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		var resultValue string
		switch mType {
		case "gauge":
			resultValue = fmt.Sprintf("%v", *metric.Value)
		case "counter":
			resultValue = fmt.Sprintf("%v", *metric.Delta)
		default:
			http.Error(w, "no such metric", http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte(resultValue))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
}

// UpdateFromJSON godoc
//
//	@Accept			application/json
//	@Produce		application/json
//	@Success		200		{string}	json	"OK"
//	@Failure		400		{string}	error	"Bad Request"
//	@Failure		404		{string}	error	"Not Found"
//	@Router			/update [POST]
func (v *APIV1) UpdateFromJSON() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var metric models.Metric
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(&metric)
		if err != nil {
			logger.Log.Debug("cannot decode json request body", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		outMetric, err := v.Storage.SaveMetric(r.Context(), metric)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		enc := json.NewEncoder(w)
		if err := enc.Encode(&outMetric); err != nil {
			logger.Log.Debug("cannot encode json request body", zap.Error(err))
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		}
	}
}

// Update godoc
//
//	@Accept			application/json
//	@Produce		application/json
//	@Success		200		{string}	json	"OK"
//	@Failure		400		{string}	error	"Bad Request"
//	@Failure		404		{string}	error	"Not Found"
//	@Router			/update/{metricType}/{metricName}/{metricValue} [POST]
func (v *APIV1) Update() http.HandlerFunc {
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

		var metric = models.Metric{ID: metricName, MType: metricType}

		switch metricType {
		case "gauge":
			value, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				http.Error(w, "Metric value is wrong", http.StatusBadRequest)
				return
			}
			metric.Value = &value
		case "counter":
			delta, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				http.Error(w, "Metric value is wrong", http.StatusBadRequest)
				return
			}
			metric.Delta = &delta
		default:
			http.Error(w, "No such metric type", http.StatusNotFound)
			return
		}

		outMetric, err := v.Storage.SaveMetric(r.Context(), metric)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		enc := json.NewEncoder(w)
		if err := enc.Encode(&outMetric); err != nil {
			logger.Log.Debug("cannot encode json request body", zap.Error(err))
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		}
	}
}

// BatchUpdate godoc
//
//	@Accept			application/json
//	@Produce		application/json
//	@Success		200		{string}	json	"OK"
//	@Failure		404		{string}	error	"Not Found"
//	@Failure		404		{string}	error	"Internal Server Error"
//	@Router			/updates [POST]
func (v *APIV1) BatchUpdate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var metrics []models.Metric
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&metrics); err != nil {
			logger.Log.Debug("cannot decode json request body", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		outMetrics, err := v.Storage.SaveBatch(r.Context(), metrics)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Encoding", "gzip")
		w.WriteHeader(http.StatusOK)

		byteData, err := json.Marshal(outMetrics)
		if config.Options.Key != "" {
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

			signHeader, err := sign.GetHmacString(byteData, config.Options.Key)
			if err != nil {
				log.Printf("Err singing body with err %+v", err)
				return
			}
			r.Header.Set("HashSHA256", signHeader)
		}

		if _, err = w.Write(byteData); err != nil {
			logger.Log.Debug("cannot encode json request body", zap.Error(err))
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		}
	}
}

func (v *APIV1) NotFound() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusNotFound)
	}
}

func (v *APIV1) BadRequest() http.HandlerFunc {
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
