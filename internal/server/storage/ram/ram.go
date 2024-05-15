package ram

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/aykuli/observer/internal/server/models"
	"github.com/aykuli/observer/internal/server/storage"
)

type Storage struct {
	memStorage storage.MemStorage
	mutex      sync.RWMutex
}

func NewStorage() (*Storage, error) {
	var rs = Storage{memStorage: storage.MemStorage{
		GaugeMetrics:   map[string]float64{},
		CounterMetrics: map[string]int64{},
	}}
	return &rs, nil
}
func (rs *Storage) Ping(ctx context.Context) error {
	return nil
}

func (rs *Storage) GetMetrics(ctx context.Context) (string, error) {
	var metrics []string
	rs.mutex.RLock()
	for k, v := range rs.memStorage.GaugeMetrics {
		metrics = append(metrics, fmt.Sprintf("%s: %f", k, v))
	}
	for k, d := range rs.memStorage.CounterMetrics {
		metrics = append(metrics, fmt.Sprintf("%s: %d", k, d))
	}
	rs.mutex.RUnlock()

	return strings.Join(metrics, ",\n"), nil
}

func (rs *Storage) ReadMetric(ctx context.Context, mName, mType string) (*models.Metric, error) {
	outMt := models.Metric{ID: mName, MType: mType}
	var value float64
	var delta int64
	var ok bool

	rs.mutex.RLock()
	switch mType {
	case "gauge":
		value, ok = rs.memStorage.GaugeMetrics[mName]
		if !ok {
			return nil, errors.New("no such metric")
		}
		outMt.Value = &value
	case "counter":
		delta, ok = rs.memStorage.CounterMetrics[mName]
		if !ok {
			return nil, errors.New("no such metric")
		}
		outMt.Delta = &delta
	default:
		return nil, errors.New("no such metric")
	}
	rs.mutex.RUnlock()

	return &outMt, nil
}

func (rs *Storage) SaveMetric(ctx context.Context, metric models.Metric) (*models.Metric, error) {
	outMt := models.Metric{ID: metric.ID, MType: metric.MType}
	var value float64
	var delta int64

	rs.mutex.Lock()
	switch metric.MType {
	case "gauge":
		value = *metric.Value
		rs.memStorage.GaugeMetrics[metric.ID] = value
		outMt.Value = &value
	case "counter":
		delta = *metric.Delta
		rs.memStorage.CounterMetrics[metric.ID] += delta
		delta = rs.memStorage.CounterMetrics[metric.ID]
		outMt.Delta = &delta
	default:
		return nil, newRAMErr("SaveMetric", errors.New("no such metric type"))
	}
	rs.mutex.RLock()

	return &outMt, nil
}

func (rs *Storage) SaveBatch(ctx context.Context, metrics []models.Metric) ([]models.Metric, error) {
	var outMetrics []models.Metric
	var value float64
	var delta int64

	rs.mutex.Lock()
	for _, mt := range metrics {
		outMt := models.Metric{ID: mt.ID, MType: mt.MType}

		switch mt.MType {
		case "gauge":
			value = *mt.Value
			rs.memStorage.GaugeMetrics[mt.ID] = value
			outMt.Value = &value
		case "counter":
			delta = *mt.Delta
			rs.memStorage.CounterMetrics[mt.ID] += delta
			delta = rs.memStorage.CounterMetrics[mt.ID]
			outMt.Delta = &delta
		default:
			return nil, newRAMErr("SaveBatch", errors.New("no such metric type"))
		}

		outMetrics = append(outMetrics, outMt)
	}
	rs.mutex.Unlock()

	return outMetrics, nil
}
