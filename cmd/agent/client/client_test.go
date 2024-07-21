package client

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aykuli/observer/internal/agent/config"
	"github.com/aykuli/observer/internal/agent/storage"
)

func TestSendBatchMetrics(t *testing.T) {
	reqCounter := 0
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqCounter++
	}))
	defer testServer.Close()

	memstorage := storage.NewMemStorage()
	memstorage.GarbageStats()
	configAddr, ok := strings.CutPrefix(testServer.URL, "http://")
	require.True(t, ok)

	options := config.Config{
		Address:        configAddr,
		ReportInterval: 3,
		PollInterval:   2,
		Key:            "",
		RateLimit:      10,
	}

	client := NewMetricsClient(options, &memstorage)

	t.Run("sends batch", func(t *testing.T) {
		client.SendBatchMetrics()
		assert.Equal(t, 1, reqCounter)
	})

	t.Run("sends by one", func(t *testing.T) {
		client.SendMetrics()
		assert.Equal(t, len(memstorage.GetAllMetrics())+1, reqCounter)
	})
}
