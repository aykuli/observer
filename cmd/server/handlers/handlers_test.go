package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/aykuli/observer/internal/models"
	"github.com/aykuli/observer/internal/server/config"
	"github.com/aykuli/observer/internal/server/storage/local"
)

func TestHandlers(t *testing.T) {
	logger := zap.NewExample()
	defer logger.Sync()

	c := config.Config{Restore: true}
	c.Init(logger)
	storage, err := local.NewStorage(c, logger)
	require.NoError(t, err)

	r := chi.NewRouter()
	api := APIV1{Storage: storage, CryptoPrivKeyPath: c.CryptoPrivKeyPath, Key: c.Key}

	// register endpoints
	r.Get("/ping", api.Ping())
	r.Get("/", api.GetAllMetrics())
	r.Post("/update", api.UpdateFromJSON())
	r.Post("/value", api.ReadMetric())
	r.Get("/value/{metricType}/{metricName}", api.GetMetric())
	r.Post("/update/{metricType}/{metricName}/{metricValue}", api.Update())
	r.Post("/updates", api.BatchUpdate())

	testServer := httptest.NewServer(r)
	// example gauge metrics
	valueA := 5.2
	metricID := "a_test"
	metric := models.Metric{
		ID:    metricID,
		MType: "gauge",
		Delta: nil,
		Value: &valueA,
	}

	// "/ping" endpoint example
	res, err := http.Get(testServer.URL + "/ping")
	require.NoError(t, err)
	pong, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	require.Equal(t, "pong", string(pong))
	err = res.Body.Close()
	require.NoError(t, err)

	// "/update" endpoint example
	byteData, err := json.Marshal(metric)
	require.NoError(t, err)

	res, err = http.Post(testServer.URL+"/update", "application/json", bytes.NewBuffer(byteData))
	require.NoError(t, err)
	outMetric, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	require.Equal(t, "{\"id\":\"a_test\",\"type\":\"gauge\",\"value\":5.2}", string(outMetric))

	err = res.Body.Close()
	require.NoError(t, err)

	// "/value" endpoint example
	askedMetric := models.Metric{ID: metricID, MType: "gauge"}
	byteData, err = json.Marshal(askedMetric)
	require.NoError(t, err)
	res, err = http.Post(testServer.URL+"/value", "application/json", bytes.NewBuffer(byteData))
	require.NoError(t, err)
	out, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	err = res.Body.Close()
	require.NoError(t, err)
	require.Equal(t, "{\"id\":\"a_test\",\"type\":\"gauge\",\"value\":5.2}\n", string(out))

	// "/update/{metricType}/{metricName}/{metricValue}" endpoint example
	res, err = http.Post(testServer.URL+"/update/gauge/b_test/1.2", "application/json", bytes.NewBuffer([]byte("")))
	require.NoError(t, err)

	out, err = io.ReadAll(res.Body)
	require.NoError(t, err)
	err = res.Body.Close()
	require.NoError(t, err)
	require.Equal(t, "{\"id\":\"b_test\",\"type\":\"gauge\",\"value\":1.2}\n", string(out))

	// "/value/{metricType}/{metricName}" endpoint example
	res, err = http.Get(testServer.URL + "/value/gauge/" + metricID)
	require.NoError(t, err)
	out, err = io.ReadAll(res.Body)
	require.NoError(t, err)
	err = res.Body.Close()
	require.NoError(t, err)
	require.Equal(t, "5.2", string(out))

	// "/" endpoint example
	res, err = http.Get(testServer.URL)
	require.NoError(t, err)
	out, err = io.ReadAll(res.Body)
	require.NoError(t, err)
	err = res.Body.Close()
	require.NoError(t, err)
	require.Equal(t, "a_test: 5.200000,\nb_test: 1.200000,\nb_test: 5", string(out))

	// "/updates" endpoint example
	valueC := 1.3
	valueD := 1.4
	metrics := []models.Metric{
		{ID: "c_test", MType: "gauge", Delta: nil, Value: &valueC},
		{ID: "c_test", MType: "gauge", Delta: nil, Value: &valueD},
	}
	byteData, err = json.Marshal(metrics)
	require.NoError(t, err)

	res, err = http.Post(testServer.URL+"/updates", "application/json", bytes.NewBuffer(byteData))
	require.NoError(t, err)

	out, err = io.ReadAll(res.Body)
	require.NoError(t, err)
	err = res.Body.Close()
	require.NoError(t, err)
	require.Equal(t, "[{\"id\":\"c_test\",\"type\":\"gauge\",\"value\":1.3},{\"id\":\"c_test\",\"type\":\"gauge\",\"value\":1.4}]", string(out))
}
