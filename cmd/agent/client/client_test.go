package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aykuli/observer/internal/agent/config"
	"github.com/aykuli/observer/internal/agent/storage"
	"github.com/aykuli/observer/internal/logger"
)

func TestSendBatchMetrics(t *testing.T) {
	reqCounter := 0
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqCounter++
	}))
	defer testServer.Close()
	var wg sync.WaitGroup
	memstorage := storage.NewMemStorage()
	memstorage.GarbageStats(&wg)
	configAddr, ok := strings.CutPrefix(testServer.URL, "http://")
	require.True(t, ok)

	options := config.Config{
		Address:        configAddr,
		ReportInterval: 3,
		PollInterval:   2,
		Key:            "",
		RateLimit:      10,
	}
	agentLogger, err := logger.New()
	require.NoError(t, err)
	client := NewMetricsClient(options, &memstorage, agentLogger)
	ctx := context.Background()

	t.Run("sends batch", func(t *testing.T) {
		client.SendBatchMetrics(ctx, &wg)
		assert.Equal(t, 1, reqCounter)
	})

	t.Run("sends by one", func(t *testing.T) {
		client.SendMetrics(ctx, &wg)
		assert.Equal(t, len(memstorage.GetAllMetrics())+1, reqCounter)
	})
	wg.Wait()

}
