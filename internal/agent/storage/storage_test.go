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

	tests := []struct {
		name string
	}{
		{
			name: "check if it works",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			GetStats(&memStorage)

			assert.Contains(t, memStorage.CounterMetrics, "PollCount")
			assert.Contains(t, memStorage.GaugeMetrics, "LastGC")
			assert.Contains(t, memStorage.GaugeMetrics, "MSpanSys")
			assert.Contains(t, memStorage.GaugeMetrics, "StackInuse")
			assert.Contains(t, memStorage.GaugeMetrics, "StackInuse")
		})
	}

}
