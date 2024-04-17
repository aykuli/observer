package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetStats(t *testing.T) {
	memStorage := MemStorage{
		GaugeMetrics:   map[string]float64{},
		CounterMetrics: map[string]int64{},
	}

	t.Run("get gauge stats", func(t *testing.T) {
		memStorage.GetStats()

		assert.Contains(t, memStorage.CounterMetrics, "PollCount")
		assert.Contains(t, memStorage.GaugeMetrics, "LastGC")
		assert.Contains(t, memStorage.GaugeMetrics, "MSpanSys")
		assert.Contains(t, memStorage.GaugeMetrics, "StackInuse")
		assert.Contains(t, memStorage.GaugeMetrics, "StackInuse")
	})

	t.Run("get system util info values", func(t *testing.T) {
		memStorage.GetSystemUtilInfo()

		assert.Contains(t, memStorage.GaugeMetrics, "TotalMemory")
		assert.Contains(t, memStorage.GaugeMetrics, "FreeMemory")
		assert.Contains(t, memStorage.GaugeMetrics, "CPUutilization1")
	})
}
