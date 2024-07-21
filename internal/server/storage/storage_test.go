package storage

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aykuli/observer/internal/models"
	"github.com/aykuli/observer/internal/server/config"
)

func TestMemStorage(t *testing.T) {
	metricsMap := NewMetricsMap(config.Options.FileStoragePath, true)
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
