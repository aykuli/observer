// Package storage provides methods to handle metrics.
package storage

import (
	"context"
	"errors"
	"sync"

	"github.com/aykuli/observer/internal/models"
)

// Storage interface provides methods need to be provided by the Storage object.
type Storage interface {
	Ping(ctx context.Context) error
	GetMetrics(ctx context.Context) (string, error)
	ReadMetric(ctx context.Context, metricName, metricType string) (*models.Metric, error)
	SaveMetric(ctx context.Context, metric models.Metric) (*models.Metric, error)
	SaveBatch(ctx context.Context, metrics []models.Metric) ([]models.Metric, error)
}

type GaugeMetrics map[string]float64
type CounterMetrics map[string]int64

// Metrics struct keeps gauge and counter metrics
type Metrics struct {
	Gauge   GaugeMetrics   `json:"gauge_metrics"`
	Counter CounterMetrics `json:"counter_metrics"`
}

// MetricsMap struct keeps metrics and configuration on metrics handling
type MetricsMap struct {
	metrics     Metrics
	mutex       sync.RWMutex
	filepath    string
	flushOnSave bool
}

// NewMetricsMap creates MetricsMap object based on configuration provided on application start.
func NewMetricsMap(filepath string, flushOnSave bool) *MetricsMap {
	return &MetricsMap{
		metrics: Metrics{
			Gauge:   make(map[string]float64),
			Counter: make(map[string]int64),
		},
		mutex:       sync.RWMutex{},
		filepath:    filepath,
		flushOnSave: flushOnSave,
	}
}

// GetGauge returns gauge metric value.
func (ms *MetricsMap) GetGauge(mName string) (float64, bool) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()
	val, ok := ms.metrics.Gauge[mName]
	return val, ok
}

// GetCounter returns counter metric value.
func (ms *MetricsMap) GetCounter(mName string) (int64, bool) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()
	val, ok := ms.metrics.Counter[mName]
	return val, ok
}

// SaveGauge saves gauge metric value.
func (ms *MetricsMap) SaveGauge(mName string, value float64) (float64, error) {
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

// SaveCounter saves counter metric value.
func (ms *MetricsMap) SaveCounter(mName string, delta int64) (int64, error) {
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

// LoadFromFile reads metrics from file and saves it to the object.
func (ms *MetricsMap) LoadFromFile() error {
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

func (ms *MetricsMap) flushToDisk() error {
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

// SaveToFile saves metrics from the memory to file.
func (ms *MetricsMap) SaveToFile() error {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()
	return ms.flushToDisk()
}

// GetGaugeMetrics returns map only with gauge metrics.
func (ms *MetricsMap) GetGaugeMetrics() GaugeMetrics {
	return ms.metrics.Gauge
}

// GetCounterMetrics returns map only with counter metrics.
func (ms *MetricsMap) GetCounterMetrics() CounterMetrics {
	return ms.metrics.Counter
}
