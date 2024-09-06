package routers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/aykuli/observer/internal/server/config"
	"github.com/aykuli/observer/internal/server/storage/local"
)

func TestLocalStorage(t *testing.T) {
	logger := zap.NewExample()
	defer logger.Sync()

	options := config.Config{
		StoreInterval:   0,
		FileStoragePath: "",
		Restore:         false,
	}
	store, err := local.NewStorage(options, logger)
	require.NoError(t, err)

	ts := httptest.NewServer(MetricsRouter(store, logger, options))
	defer ts.Close()

	t.Run("init storage should be empty", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
		req.Header.Set("Accept-Encoding", "")

		require.NoError(t, err)
		resp, err := ts.Client().Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "", string(respBody))
	})

	type want struct {
		code     int
		respBody string
	}
	tests := []struct {
		name       string
		method     string
		requestURL string
		want
	}{
		{
			name:       "Update gauge metric value",
			method:     http.MethodPost,
			requestURL: "/update/gauge/metric0/0.5",
			want: want{code: http.StatusOK, respBody: `{"id":"metric0","type":"gauge","value":0.5}
`},
		},
		{
			name:       "Update counter metric value",
			method:     http.MethodPost,
			requestURL: "/update/counter/metric1/12",
			want: want{code: http.StatusOK, respBody: `{"id":"metric1","type":"counter","delta":12}
`},
		},
		{
			name:       "Get gauge metric current value",
			method:     http.MethodGet,
			requestURL: "/value/gauge/metric0",
			want:       want{code: http.StatusOK, respBody: `0.5`},
		},
		{
			name:       "Get counter metric current value",
			method:     http.MethodGet,
			requestURL: "/value/counter/metric1",
			want:       want{code: http.StatusOK, respBody: "12"},
		},
		{
			name:       "Get wrong metric type",
			method:     http.MethodGet,
			requestURL: "/value/unknown/metric1",
			want:       want{code: http.StatusNotFound, respBody: "no such metric\n"},
		},
		{
			name:       "Get wrong counter metric current value",
			method:     http.MethodGet,
			requestURL: "/value/counter/metric2",
			want:       want{code: http.StatusNotFound, respBody: "no such metric\n"},
		},
		{
			name:       "Get all metrics",
			method:     http.MethodGet,
			requestURL: "/",
			want: want{
				code:     http.StatusOK,
				respBody: `metric0: 0.500000`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, ts.URL+tt.requestURL, nil)
			req.Header.Set("Accept-Encoding", "")

			require.NoError(t, err)
			resp, err := ts.Client().Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			respBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Equal(t, tt.want.code, resp.StatusCode)
			if tt.method == http.MethodGet {
				assert.Contains(t, string(respBody), tt.want.respBody)
			} else if tt.method == http.MethodPost {
				assert.Equal(t, tt.want.respBody, string(respBody))
			}
		})
	}
}

func TestUpdateMetricsRouter(t *testing.T) {
	logger := zap.NewExample()
	defer logger.Sync()

	options := config.Config{
		StoreInterval:   0,
		FileStoragePath: "",
		Restore:         false,
	}
	store, err := local.NewStorage(options, logger)
	require.NoError(t, err)
	ts := httptest.NewServer(MetricsRouter(store, logger, options))
	defer ts.Close()

	type want struct{ code int }

	tests := []struct {
		name       string
		method     string
		requestURL string

		want
	}{
		{
			name:       "Update not allowed with GET method",
			method:     http.MethodGet,
			requestURL: "/update/gauge/metric0/5",
			want:       want{code: http.StatusMethodNotAllowed},
		},
		{
			name:       "Update gauge metric request",
			method:     http.MethodPost,
			requestURL: "/update/gauge/metric0/5",
			want:       want{code: http.StatusOK},
		},
		{
			name:       "Metric type is wrong",
			method:     http.MethodPost,
			requestURL: "/update/random_string/testGauge/5",
			want:       want{code: http.StatusBadRequest},
		},
		{
			name:       "Metric name is empty",
			method:     http.MethodPost,
			requestURL: "/update/gauge",
			want:       want{code: http.StatusNotFound},
		},
		{
			name:       "Gauge metric value is wrong",
			method:     http.MethodPost,
			requestURL: "/update/gauge/metric0/random",
			want:       want{code: http.StatusBadRequest},
		},
		{
			name:       "Counter metric value is wrong",
			method:     http.MethodPost,
			requestURL: "/update/counter/metric1/56.789",
			want:       want{code: http.StatusBadRequest},
		},
		{
			name:       "Gauge metric value is valid",
			method:     http.MethodPost,
			requestURL: "/update/gauge/metric0/56.789",
			want:       want{code: http.StatusOK}},
		{
			name:       "Counter metric value is valid",
			method:     http.MethodPost,
			requestURL: "/update/counter/metric1/56",
			want:       want{code: http.StatusOK},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, ts.URL+tt.requestURL, nil)
			require.NoError(t, err)

			resp, err := ts.Client().Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			_, err = io.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Equal(t, tt.want.code, resp.StatusCode)
		})
	}
}
