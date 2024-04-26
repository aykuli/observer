package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGarbageStats(t *testing.T) {
	t.Run("Check stats collection", func(t *testing.T) {
		memStorage := NewMemStorage()
		memStorage.GarbageStats()

		var metricNames []string
		for _, mt := range memStorage.GetAllMetrics() {
			metricNames = append(metricNames, mt.ID)
		}

		assert.Contains(t, metricNames, "PollCount")
		assert.Contains(t, metricNames, "LastGC")
		assert.Contains(t, metricNames, "MSpanSys")
		assert.Contains(t, metricNames, "StackInuse")
		assert.Contains(t, metricNames, "StackInuse")
	})

}
