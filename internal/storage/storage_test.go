package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetStats(t *testing.T) {
	memstorage := MemStorage{
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
			GetStats(&memstorage)

			assert.Contains(t, memstorage.CounterMetrics, "PollCount")
			assert.Contains(t, memstorage.GaugeMetrics, "LastGC")
			assert.Contains(t, memstorage.GaugeMetrics, "MSpanSys")
			assert.Contains(t, memstorage.GaugeMetrics, "StackInuse")
			assert.Contains(t, memstorage.GaugeMetrics, "StackInuse")
		})
	}

}
