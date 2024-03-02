package routers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aykuli/observer/internal/server/storage"
)

func TestGetMetricsRouter(t *testing.T) {
	type want struct {
		code     int
		respBody string
	}
	tests := []struct {
		name       string
		method     string
		requestURL string
		memStorage storage.MemStorage
		want
	}{
		{
			name:       "GET ",
			method:     http.MethodGet,
			requestURL: "/",
			memStorage: storage.MemStorage{
				GaugeMetrics:   map[string]float64{"metric0": 0.5},
				CounterMetrics: map[string]int64{"metric1": 12},
			},
			want: want{
				code:     http.StatusOK,
				respBody: "metric1 metric0",
			},
		},
		{
			name:       "Get gauge metric current value",
			method:     http.MethodGet,
			requestURL: "/value/gauge/metric0",
			memStorage: storage.MemStorage{
				GaugeMetrics:   map[string]float64{"metric0": 0.5},
				CounterMetrics: map[string]int64{"metric1": 12},
			},
			want: want{code: http.StatusOK, respBody: "0.5"},
		},
		{
			name:       "Get counter metric current value",
			method:     http.MethodGet,
			requestURL: "/value/counter/metric1",
			memStorage: storage.MemStorage{
				GaugeMetrics:   map[string]float64{"metric0": 0.5},
				CounterMetrics: map[string]int64{"metric1": 12},
			},
			want: want{code: http.StatusOK, respBody: "12"},
		},
		{
			name:       "Get wrong metric type",
			method:     http.MethodGet,
			requestURL: "/value/unknown/metric1",
			memStorage: storage.MemStorage{
				GaugeMetrics:   map[string]float64{"metric0": 0.5},
				CounterMetrics: map[string]int64{"metric1": 12},
			},
			want: want{code: http.StatusNotFound, respBody: "No such metric\n"},
		},
		{
			name:       "Get wrong counter metric current value",
			method:     http.MethodGet,
			requestURL: "/value/counter/metric2",
			memStorage: storage.MemStorage{
				GaugeMetrics:   map[string]float64{"metric0": 0.5},
				CounterMetrics: map[string]int64{"metric1": 12},
			},
			want: want{code: http.StatusNotFound, respBody: "No such metric\n"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(MetricsRouter(&tt.memStorage))
			defer ts.Close()

			req, err := http.NewRequest(tt.method, ts.URL+tt.requestURL, nil)
			require.NoError(t, err)

			resp, err := ts.Client().Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			respBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Equal(t, tt.want.code, resp.StatusCode)
			assert.Equal(t, tt.want.respBody, string(respBody))
		})
	}
}

func TestUpdateMetricsRouter(t *testing.T) {
	type want struct {
		code       int
		memStorage storage.MemStorage
	}

	tests := []struct {
		name       string
		method     string
		requestURL string
		memStorage storage.MemStorage

		want
	}{
		{
			name:       "Update not allowed with GET method",
			method:     http.MethodGet,
			requestURL: "/update/gauge/metric0/5",
			memStorage: storage.MemStorage{
				GaugeMetrics:   map[string]float64{"metric0": 0.5},
				CounterMetrics: map[string]int64{"metric1": 12},
			},
			want: want{
				code: http.StatusMethodNotAllowed,
				memStorage: storage.MemStorage{
					GaugeMetrics:   map[string]float64{"metric0": 0.5},
					CounterMetrics: map[string]int64{"metric1": 12},
				},
			},
		},
		{
			name:       "Update gauge metric request",
			method:     http.MethodPost,
			requestURL: "/update/gauge/metric0/5",
			memStorage: storage.MemStorage{
				GaugeMetrics:   map[string]float64{"metric0": 0.5},
				CounterMetrics: map[string]int64{"metric1": 12},
			},
			want: want{code: http.StatusOK,

				memStorage: storage.MemStorage{
					GaugeMetrics:   map[string]float64{"metric0": 5},
					CounterMetrics: map[string]int64{"metric1": 12},
				}},
		},
		{
			name:       "Metric type is wrong",
			method:     http.MethodPost,
			requestURL: "/update/random_string/testGauge/5",
			memStorage: storage.MemStorage{
				GaugeMetrics:   map[string]float64{"metric0": 0.5},
				CounterMetrics: map[string]int64{"metric1": 12},
			},
			want: want{code: http.StatusBadRequest,
				memStorage: storage.MemStorage{
					GaugeMetrics:   map[string]float64{"metric0": 0.5},
					CounterMetrics: map[string]int64{"metric1": 12},
				},
			},
		},
		{
			name:       "Metric name is empty",
			method:     http.MethodPost,
			requestURL: "/update/gauge", memStorage: storage.MemStorage{
				GaugeMetrics:   map[string]float64{"metric0": 0.5},
				CounterMetrics: map[string]int64{"metric1": 12},
			},
			want: want{
				code: http.StatusNotFound,
				memStorage: storage.MemStorage{
					GaugeMetrics:   map[string]float64{"metric0": 0.5},
					CounterMetrics: map[string]int64{"metric1": 12},
				},
			},
		},
		{
			name:       "Gauge metric value is wrong",
			method:     http.MethodPost,
			requestURL: "/update/gauge/metric0/random",
			memStorage: storage.MemStorage{
				GaugeMetrics:   map[string]float64{"metric0": 0.5},
				CounterMetrics: map[string]int64{"metric1": 12},
			},
			want: want{
				code: http.StatusBadRequest,
				memStorage: storage.MemStorage{
					GaugeMetrics:   map[string]float64{"metric0": 0.5},
					CounterMetrics: map[string]int64{"metric1": 12},
				},
			},
		},
		{
			name:       "Counter metric value is wrong",
			method:     http.MethodPost,
			requestURL: "/update/counter/metric1/56.789",
			memStorage: storage.MemStorage{
				GaugeMetrics:   map[string]float64{"metric0": 0.5},
				CounterMetrics: map[string]int64{"metric1": 12},
			},
			want: want{
				code: http.StatusBadRequest,
				memStorage: storage.MemStorage{
					GaugeMetrics:   map[string]float64{"metric0": 0.5},
					CounterMetrics: map[string]int64{"metric1": 12},
				},
			},
		},
		{
			name:       "Gauge metric value is valid",
			method:     http.MethodPost,
			requestURL: "/update/gauge/metric0/56.789",
			memStorage: storage.MemStorage{
				GaugeMetrics:   map[string]float64{"metric0": 0.5},
				CounterMetrics: map[string]int64{"metric1": 12},
			},
			want: want{
				code: http.StatusOK,
				memStorage: storage.MemStorage{
					GaugeMetrics:   map[string]float64{"metric0": 56.789},
					CounterMetrics: map[string]int64{"metric1": 12},
				}},
		},
		{
			name:       "Counter metric value is valid",
			method:     http.MethodPost,
			requestURL: "/update/counter/metric1/56",
			memStorage: storage.MemStorage{
				GaugeMetrics:   map[string]float64{"metric0": 0.5},
				CounterMetrics: map[string]int64{"metric1": 12},
			},
			want: want{
				code: http.StatusOK,
				memStorage: storage.MemStorage{
					GaugeMetrics:   map[string]float64{"metric0": 0.5},
					CounterMetrics: map[string]int64{"metric1": 68},
				}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(MetricsRouter(&tt.memStorage))
			defer ts.Close()

			req, err := http.NewRequest(tt.method, ts.URL+tt.requestURL, nil)
			require.NoError(t, err)

			resp, err := ts.Client().Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			_, err = io.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Equal(t, tt.want.code, resp.StatusCode)
			assert.Equal(t, tt.want.memStorage, tt.memStorage)

		})
	}
}
