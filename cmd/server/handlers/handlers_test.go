package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aykuli/observer/internal/storage"
)

func TestCheckType(t *testing.T) {
	tests := []struct {
		name  string
		value string
		valid bool
	}{
		{
			name:  "gauge test",
			value: "gauge",
			valid: true,
		},
		{
			name:  "counter test",
			value: "counter",
			valid: true,
		},
		{
			name:  "wrong string",
			value: "random string",
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, checkType(tt.value), tt.valid)
		})
	}
}

func TestUpdateRuntime(t *testing.T) {
	type want struct {
		code int
	}

	tests := []struct {
		name       string
		method     string
		requestURL string
		metricType string
		metricName string

		want
	}{
		{
			name:       "GET request",
			method:     http.MethodGet,
			requestURL: "http://localhost:8080/update/gauge/testGauge/5",
			want:       want{code: http.StatusMethodNotAllowed},
		},
		{
			name:       "Metric type is wrong",
			method:     http.MethodPost,
			requestURL: "http://localhost:8080/update/random_string/testGauge/5",
			want:       want{code: http.StatusBadRequest},
		},
		{
			name:       "Metric name is empty",
			method:     http.MethodPost,
			requestURL: "http://localhost:8080/update/gauge",
			want:       want{code: http.StatusNotFound},
		},
		{
			name:       "Gauge metric value is wrong",
			method:     http.MethodPost,
			requestURL: "http://localhost:8080/update/gauge/weirdos/testGauge/8",
			want:       want{code: http.StatusBadRequest},
		},
		{
			name:       "Counter metric value is wrong",
			method:     http.MethodPost,
			requestURL: "http://localhost:8080/update/counter/testCounter/56.789",
			want:       want{code: http.StatusBadRequest},
		},
		{
			name:       "Gauge metric value is valid",
			method:     http.MethodPost,
			requestURL: "http://localhost:8080/update/gauge/testGauge/56.789",
			metricType: "gauge",
			metricName: "testGauge",
			want:       want{code: http.StatusOK},
		},
		{
			name:       "Counter metric value is valid",
			method:     http.MethodPost,
			requestURL: "http://localhost:8080/update/counter/testCounter/56",
			metricType: "counter",
			metricName: "testCounter",
			want:       want{code: http.StatusOK},
		},
	}

	memstorage := storage.MemStorage{
		GaugeMetrics:   map[string]float64{},
		CounterMetrics: map[string]int64{},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, tt.requestURL, nil)
			w := httptest.NewRecorder()
			UpdateRuntime(memstorage)(w, request)

			res := w.Result()
			defer res.Body.Close()
			assert.Equal(t, tt.want.code, res.StatusCode)

			if tt.want.code == http.StatusOK {
				var s interface{}
				if tt.metricType == "gauge" {
					s = memstorage.GaugeMetrics
				}
				if tt.metricType == "counter" {
					s = memstorage.CounterMetrics
				}
				assert.Contains(t, s, tt.metricName)
			}
		})
	}
}
