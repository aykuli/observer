package client

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"

	"github.com/aykuli/observer/internal/agent/storage"
)

func TestSendMetrics(t *testing.T) {
	t.Run("check if works", func(t *testing.T) {
		reqCounter := 0

		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqCounter++
		}))
		defer testServer.Close()
		req := resty.New().R()

		memstorage := storage.NewMemStorage()
		memstorage.GarbageStats()
		client := NewMetricsClint(testServer.URL, memstorage)

		client.SendMetrics(req)
		client.SendBatchMetrics(req)
		metrics := memstorage.GetAllMetrics()

		assert.Equal(t, len(metrics)+1, reqCounter)
	})
}
