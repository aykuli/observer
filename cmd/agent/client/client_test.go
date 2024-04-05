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
	tests := []struct {
		name       string
		memStorage storage.MemStorage
	}{
		{
			name: "gauge and count metrics sent",
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

			newClient := MerticsClient{ServerAddr: testServer.URL, MemStorage: tt.memStorage}
			newClient.SendMetrics(req)

			assert.Equal(t, len(tt.memStorage.GaugeMetrics)+len(tt.memStorage.CounterMetrics), reqCounter)
		})

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

			newClient := MerticsClient{ServerAddr: testServer.URL, MemStorage: tt.memStorage}
			newClient.SendBatchMetrics(req)
			reqCount := 0
			if len(tt.memStorage.CounterMetrics) > 0 || len(tt.memStorage.GaugeMetrics) > 0 {
				reqCount = 1
			}
			assert.Equal(t, reqCount, reqCounter)
		})
	}
}
