// Package handlers provides methods handling endpoints.
package handlers

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/aykuli/observer/internal/models"
	"github.com/aykuli/observer/internal/server/config"
	"github.com/aykuli/observer/internal/server/storage"
	"github.com/aykuli/observer/internal/sign"
)

// APIV1 struct keeps storage struct and provides methods for endpoints routing
type APIV1 struct {
	Storage storage.Storage
	Logger  zap.SugaredLogger
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
			v.Logger.Errorln("ping storage error", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte("pong"))
		if err != nil {
			v.Logger.Errorln("response body writing error", zap.Error(err))
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		r.Body.Close()
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
//	@Produce		text/plain
//	@Success		200		{string}	json	"OK"
//	@Failure		400		{string}	error	"Bad Request"
//	@Failure		422		{string}	error	"Unprocessable Entity"
//	@Failure		500		{string}	error	"Internal Server Error"
//	@Router			/value [POST]
func (v *APIV1) ReadMetric() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var askedMetric models.Metric
		if err := json.NewDecoder(r.Body).Decode(&askedMetric); err != nil {
			v.Logger.Errorln("cannot decode json request body", zap.Error(err))
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
		if err := json.NewEncoder(w).Encode(&metric); err != nil {
			v.Logger.Errorln("cannot encode json request body", zap.Error(err))
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		}

		r.Body.Close()
	}
}

// GetMetric godoc
//
//	@Accept			application/json
//	@Produce		text/plain
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
//	@Success		200		{object}	models.Metric	"OK"
//	@Failure		400		{string}	error	        "Bad Request"
//	@Failure		404		{string}	error	        "Not Found"
//	@Router			/update [POST]
func (v *APIV1) UpdateFromJSON() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var metric models.Metric
		if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
			v.Logger.Errorln("cannot decode json request body", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		equal := sign.Verify(metric, config.Options.Key, r.Header.Get("HashSHA256"))
		if !equal {
			v.Logger.Errorln("signs not equal", zap.Error(errors.New("signs not equal")))
			http.Error(w, "cannot serve this agent", http.StatusBadRequest)
			return
		}

		outMetric, err := v.Storage.SaveMetric(r.Context(), metric)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		byteData, err := json.Marshal(outMetric)
		if err != nil {
			v.Logger.Errorln("cannot marshal metrics", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if config.Options.Key != "" {
			w.Header().Set("HashSHA256", sign.GetHmacString(byteData, config.Options.Key))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if _, err = w.Write(byteData); err != nil {
			v.Logger.Errorln("cannot encode json request body", zap.Error(err))
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		}
	}
}

// Update godoc
//
//	@Accept			application/json
//	@Produce		application/json
//	@Success		200		{object}	models.Metric	"OK"
//	@Failure		400		{string}	error	        "Bad Request"
//	@Failure		404		{string}	error	        "Not Found"
//	@Router			/update/{metricType}/{metricName}/{metricValue} [POST]
func (v *APIV1) Update() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
			v.Logger.Errorln("cannot encode json request body", zap.Error(err))
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		}
	}
}

// BatchUpdate godoc
//
//	@Accept			application/json
//	@Produce		application/json
//	@Success		200		{array}	  models.Metric	"OK"
//	@Failure		404		{string}	error	"Not Found"
//	@Failure		404		{string}	error	"Internal Server Error"
//	@Router			/updates [POST]
func (v *APIV1) BatchUpdate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var metrics []models.Metric
		if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
			v.Logger.Errorln("cannot decode json request body", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		equal := sign.Verify(metrics, config.Options.Key, r.Header.Get("HashSHA256"))
		if !equal {
			v.Logger.Errorln("signs not equal", zap.Error(errors.New("signs not equal")))
			http.Error(w, "cannot serve this agent", http.StatusBadRequest)
			return
		}

		outMetrics, err := v.Storage.SaveBatch(r.Context(), metrics)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		slices.SortFunc(outMetrics, func(a, b models.Metric) int {
			return cmp.Compare(a.ID, b.ID)
		})

		body, err := json.Marshal(outMetrics)
		if err != nil {
			v.Logger.Errorln("cannot marshal metrics", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if config.Options.Key != "" {
			w.Header().Set("HashSHA256", sign.GetHmacString(body, config.Options.Key))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err = w.Write(body); err != nil {
			v.Logger.Errorln("cannot encode json request body", zap.Error(err))
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		}
	}
}

func checkType(metricType string) bool {
	if metricType == "gauge" || metricType == "counter" {
		return true
	}

	return false
}
