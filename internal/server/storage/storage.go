package storage

import (
	"context"
	"errors"
	"sync"

	"github.com/aykuli/observer/internal/server/models"
)

type GaugeMetrics map[string]float64
type CounterMetrics map[string]int64
type Metrics struct {
	Gauge   GaugeMetrics   `json:"gauge_metrics"`
	Counter CounterMetrics `json:"counter_metrics"`
}

type MemStorage struct {
	metrics     Metrics
	mutex       sync.RWMutex
	filepath    string
	flushOnSave bool
}

func NewMemStorage(filepath string, flushOnSave bool) MemStorage {
	return MemStorage{
		metrics: Metrics{
			Gauge:   make(map[string]float64),
			Counter: make(map[string]int64),
		},
		mutex:       sync.RWMutex{},
		filepath:    filepath,
		flushOnSave: flushOnSave,
	}
}

func (ms *MemStorage) GetGauge(mName string) (float64, bool) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()
	val, ok := ms.metrics.Gauge[mName]
	return val, ok
}

func (ms *MemStorage) GetCounter(mName string) (int64, bool) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()
	val, ok := ms.metrics.Counter[mName]
	return val, ok
}

func (ms *MemStorage) SaveGauge(mName string, value float64) (float64, error) {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	ms.metrics.Gauge[mName] = value

	if ms.flushOnSave {
		if err := ms.flushToDisk(); err != nil {
			return value, err
		}
	}

	return ms.metrics.Gauge[mName], nil
}

func (ms *MemStorage) SaveCounter(mName string, delta int64) (int64, error) {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	ms.metrics.Counter[mName] += delta

	if ms.flushOnSave {
		if err := ms.flushToDisk(); err != nil {
			return delta, err
		}
	}

	return ms.metrics.Counter[mName], nil
}

func (ms *MemStorage) LoadFromFile() error {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	consumer, err := NewConsumer(ms.filepath)
	if err != nil {
		return nil
	}

	fStore, err := consumer.ReadMetrics()
	if err != nil && errors.Is(err, ErrNoData) {
		ms.metrics = Metrics{
			Gauge:   make(map[string]float64),
			Counter: make(map[string]int64),
		}
		return nil
	}

	if err != nil {
		return err
	}

	ms.metrics = fStore
	return nil
}

func (ms *MemStorage) flushToDisk() error {
	producer, err := NewProducer(ms.filepath)
	if err != nil {
		return err
	}
	defer producer.Close()

	if len(ms.metrics.Gauge) > 0 || len(ms.metrics.Counter) > 0 {
		if err := producer.WriteMetrics(ms.metrics); err != nil {
			return err
		}
	}

	return nil
}

func (ms *MemStorage) SaveToFile() error {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()
	return ms.flushToDisk()
}

func (ms *MemStorage) GetGaugeMetrics() GaugeMetrics {
	return ms.metrics.Gauge
}

func (ms *MemStorage) GetCounterMetrics() CounterMetrics {
	return ms.metrics.Counter
}

type Storage interface {
	Ping(ctx context.Context) error
	GetMetrics(ctx context.Context) (string, error)
	ReadMetric(ctx context.Context, metricName, metricType string) (*models.Metric, error)
	SaveMetric(ctx context.Context, metric models.Metric) (*models.Metric, error)
	SaveBatch(ctx context.Context, metrics []models.Metric) ([]models.Metric, error)
}
