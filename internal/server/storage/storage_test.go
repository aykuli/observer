package storage

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/aykuli/observer/internal/models"
	"github.com/aykuli/observer/internal/server/config"
)

func TestMemStorage(t *testing.T) {
	logger := zap.NewExample()
	defer logger.Sync()

	var options config.Config
	options.Init(logger)
	metricsMap := NewMetricsMap(options.FileStoragePath, true)
	require.Empty(t, metricsMap.metrics.Counter)
	require.Empty(t, metricsMap.metrics.Gauge)

	valueA := 2.1
	metricA := models.Metric{ID: "a_test"}
	outValue, err := metricsMap.SaveGauge(metricA.ID, valueA)
	require.NoError(t, err)
	require.Equal(t, outValue, valueA)

	gaugeMetrics := metricsMap.GetGaugeMetrics()
	require.InDeltaMapValues(t, gaugeMetrics, GaugeMetrics{metricA.ID: valueA}, 0)

	deltaB := int64(5)
	metricB := models.Metric{ID: "b_test"}
	outDelta, err := metricsMap.SaveCounter(metricB.ID, deltaB)
	require.NoError(t, err)
	require.Equal(t, outDelta, deltaB)

	counterMetrics := metricsMap.GetCounterMetrics()
	require.InDeltaMapValues(t, counterMetrics, CounterMetrics{metricB.ID: deltaB}, 0)

	outGauge, ok := metricsMap.GetGauge("a_test")
	require.True(t, ok)
	require.Equal(t, valueA, outGauge)

	outCounter, ok := metricsMap.GetCounter("b_test")
	require.True(t, ok)
	require.Equal(t, deltaB, outCounter)
}

func TestNewProducer(t *testing.T) {
	fname := "testFile"
	err := os.WriteFile(fname, []byte(fname), 0644)
	require.NoError(t, err)

	producer, err := NewProducer(fname)
	require.NoError(t, err)
	require.NotNil(t, producer.file)
	require.NotNil(t, producer.writer)

	err = os.Remove(fname)
	require.NoError(t, err)
	require.NoFileExists(t, fname)
}
