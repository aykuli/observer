package storage

import (
	"context"
	"sync"

	"github.com/aykuli/observer/internal/server/models"
)

type GaugeMetrics map[string]float64
type CounterMetrics map[string]int64

type MemStorage struct {
	GaugeMetrics   GaugeMetrics   `json:"gauge_metrics"`
	CounterMetrics CounterMetrics `json:"counter_metrics"`
	mutex          sync.RWMutex
}

func (ms *MemStorage) GetGauge(mName string) (float64, bool) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()
	val, ok := ms.GaugeMetrics[mName]
	return val, ok
}

func (ms *MemStorage) GetCounter(mName string) (int64, bool) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()
	val, ok := ms.CounterMetrics[mName]
	return val, ok
}

func (ms *MemStorage) SaveGauge(mName string, value float64) float64 {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()
	ms.GaugeMetrics[mName] = value
	return ms.GaugeMetrics[mName]
}

func (ms *MemStorage) SaveCounter(mName string, delta int64) int64 {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()
	ms.CounterMetrics[mName] += delta
	return ms.CounterMetrics[mName]
}

func (ms *MemStorage) GetGaugeMetrics() GaugeMetrics {
	return ms.GaugeMetrics
}

func (ms *MemStorage) GetCounterMetrics() CounterMetrics {
	return ms.CounterMetrics
}

type Storage interface {
	Ping(ctx context.Context) error
	GetMetrics(ctx context.Context) (string, error)
	ReadMetric(ctx context.Context, metricName, metricType string) (*models.Metric, error)
	SaveMetric(ctx context.Context, metric models.Metric) (*models.Metric, error)
	SaveBatch(ctx context.Context, metrics []models.Metric) ([]models.Metric, error)
}
