package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aykuli/observer/internal/storage"
)

func TestPost(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}))
	defer testServer.Close()

	tests := []struct {
		name    string
		options Options
	}{
		{
			name:    "gauge metric url",
			options: Options{testServer.URL, "gauge", "metric0", fmt.Sprintf("%v", 56.6)},
		},
		{
			name:    "count metric url",
			options: Options{testServer.URL, "counter", "metric1", fmt.Sprintf("%v", 5897)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := resty.New().R()
			req.URL = testServer.URL
			err := post(req, tt.options)
			require.NoError(t, err)
		})
	}
}

func TestSendPost(t *testing.T) {

	tests := []struct {
		name       string
		memStorage storage.MemStorage
	}{
		{
			name: "gauge and count metrics sended",
			memStorage: storage.MemStorage{
				GaugeMetrics: map[string]float64{
					"GKey1": 1.25, "GKey2": 2.25, "GKey3": 3.25},
				CounterMetrics: map[string]int64{"CKey1": 15},
			},
		},
		{
			name: "no metrics to send",
			memStorage: storage.MemStorage{
				GaugeMetrics:   map[string]float64{},
				CounterMetrics: map[string]int64{},
			},
		},
	}

	for _, tt := range tests {
		reqCounter := 0
		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqCounter++
		}))
		defer testServer.Close()

		t.Run(tt.name, func(t *testing.T) {
			req := resty.New().R()
			req.Method = http.MethodPost
			req.URL = testServer.URL

			SendPost(req, testServer.URL, tt.memStorage)

			assert.Equal(t, len(tt.memStorage.GaugeMetrics)+len(tt.memStorage.CounterMetrics), reqCounter)
		})
	}
}
