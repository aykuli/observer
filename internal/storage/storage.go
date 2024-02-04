package storage

type GaugeMetrics map[string]float64
type CounterMetrics map[string]int64

type MemStorage struct {
	GaugeMetrics
	CounterMetrics
}
