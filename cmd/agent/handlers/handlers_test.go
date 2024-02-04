package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aykuli/observer/internal/storage"
)

func TestGetUrl(t *testing.T) {
	tests := []struct {
		name        string
		metricType  string
		metricName  string
		metricValue interface{}
		urlBase     string
		wantURL     string
	}{
		{
			name:        "gauge metric url",
			metricType:  "gauge",
			metricName:  "metric0",
			metricValue: 56.6,
			urlBase:     "http://localhost:8080",
			wantURL:     "http://localhost:8080/update/gauge/metric0/56.6",
		},
		{
			name:        "count metric url",
			metricType:  "count",
			metricName:  "metric1",
			metricValue: 4569,
			urlBase:     "http://localhost:8080",
			wantURL:     "http://localhost:8080/update/count/metric1/4569",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resUrls := getURL(tt.urlBase, tt.metricType, tt.metricName, tt.metricValue)
			assert.Equal(t, tt.wantURL, resUrls)
		})
	}
}

func TestSendPost(t *testing.T) {

	tests := []struct {
		name       string
		urlBase    string
		memstorage storage.MemStorage
	}{
		{
			name:    "gauge and count metrics sended",
			urlBase: "http://localhost",
			memstorage: storage.MemStorage{
				GaugeMetrics: map[string]float64{
					"GKey1": 1.25, "GKey2": 2.25, "GKey3": 3.25},
				CounterMetrics: map[string]int64{"CKey1": 15},
			},
		},
		{
			name:    "no metrics to send",
			urlBase: "http://localhost",
			memstorage: storage.MemStorage{
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
		client := testServer.Client()

		t.Run(tt.name, func(t *testing.T) {
			SendPost(client, testServer.URL, tt.memstorage)
			assert.Equal(t, len(tt.memstorage.GaugeMetrics)+len(tt.memstorage.CounterMetrics), reqCounter)
		})
	}
}
