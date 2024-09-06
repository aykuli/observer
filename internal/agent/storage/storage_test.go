package storage

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGarbageStats(t *testing.T) {
	t.Run("Check stats collection", func(t *testing.T) {
		var wg sync.WaitGroup
		memStorage := NewMemStorage()
		memStorage.GarbageStats(&wg)

		var metricNames []string
		for _, mt := range memStorage.GetAllMetrics() {
			metricNames = append(metricNames, mt.ID)
		}
		wg.Wait()
		assert.Contains(t, metricNames, "PollCount")
		assert.Contains(t, metricNames, "LastGC")
		assert.Contains(t, metricNames, "MSpanSys")
		assert.Contains(t, metricNames, "StackInuse")
		assert.Contains(t, metricNames, "StackInuse")
	})

	t.Run("get system util info values", func(t *testing.T) {
		var wg sync.WaitGroup
		memStorage := NewMemStorage()
		memStorage.GetSystemUtilInfo(&wg)

		var metricNames []string
		for _, mt := range memStorage.GetAllMetrics() {
			metricNames = append(metricNames, mt.ID)
		}
		wg.Wait()

		assert.Contains(t, metricNames, "TotalMemory")
		assert.Contains(t, metricNames, "FreeMemory")
		assert.Contains(t, metricNames, "CPUutilization1")
	})
}
