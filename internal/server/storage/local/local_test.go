package local

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aykuli/observer/internal/models"
	"github.com/aykuli/observer/internal/server/config"
)

func TestFileStorage(t *testing.T) {
	options := config.Config{
		StoreInterval:   10,
		FileStoragePath: "1.json",
		Restore:         false,
	}
	store, err := NewStorage(options)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("Ping", func(t *testing.T) {
		err = store.Ping(ctx)
		require.NoError(t, err)
	})

	t.Run("SaveMetric gauge metric", func(t *testing.T) {
		value := 456.2
		mIn := models.Metric{
			ID:    "gmetric",
			MType: "gauge",
			Delta: nil,
			Value: &value,
		}
		metric, err := store.SaveMetric(ctx, mIn)
		require.NoError(t, err)

		require.Equal(t, *metric.Value, *mIn.Value)
		require.FileExists(t, options.FileStoragePath)

		defer os.Remove(options.FileStoragePath)
	})

	t.Run("SaveMetric counter metric", func(t *testing.T) {
		delta := int64(78)
		mIn := models.Metric{
			ID:    "rand",
			MType: "counter",
			Delta: &delta,
			Value: nil,
		}
		metric, err := store.SaveMetric(ctx, mIn)
		require.NoError(t, err)

		require.Equal(t, *metric.Delta, *mIn.Delta)
	})

	t.Run("GetMetrics", func(t *testing.T) {
		metrics, err := store.GetMetrics(ctx)
		require.NoError(t, err)

		require.Equal(t, metrics, "gmetric: 456.200000,\nrand: 78")
	})

	t.Run("SaveBatch", func(t *testing.T) {
		delta := int64(154498869)
		delta2 := int64(256)
		value := 85.63
		value2 := 18.12
		inputMetrics := []models.Metric{
			{ID: "c", MType: "counter", Delta: &delta, Value: nil},
			{ID: "c2", MType: "counter", Delta: &delta2, Value: nil},
			{ID: "v", MType: "gauge", Delta: nil, Value: &value},
			{ID: "v2", MType: "gauge", Delta: nil, Value: &value2},
		}

		_, err := store.SaveBatch(ctx, inputMetrics)
		require.NoError(t, err)

		metrics, err := store.GetMetrics(ctx)
		require.NoError(t, err)

		require.Contains(t, metrics, "gmetric: 456.200000")
		require.Contains(t, metrics, "v: 85.630000")
		require.Contains(t, metrics, "v2: 18.120000")
		require.Contains(t, metrics, "rand: 78")
		require.Contains(t, metrics, "c2: 256")
	})
}
