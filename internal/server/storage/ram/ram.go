package ram

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aykuli/observer/internal/server/models"
	"github.com/aykuli/observer/internal/server/storage"
)

type Storage struct {
	memStorage storage.MemStorage
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
	for k, v := range rs.memStorage.GetGaugeMetrics() {
		metrics = append(metrics, fmt.Sprintf("%s: %f", k, v))
	}
	for k, d := range rs.memStorage.GetCounterMetrics() {
		metrics = append(metrics, fmt.Sprintf("%s: %d", k, d))
	}

	return strings.Join(metrics, ",\n"), nil
}

func (rs *Storage) ReadMetric(ctx context.Context, mName, mType string) (*models.Metric, error) {
	outMt := models.Metric{ID: mName, MType: mType}
	var value float64
	var delta int64
	var ok bool

	switch mType {
	case "gauge":
		value, ok = rs.memStorage.GetGauge(mName)
		if !ok {
			return nil, errors.New("no such metric")
		}
		outMt.Value = &value
	case "counter":
		delta, ok = rs.memStorage.GetCounter(mName)
		if !ok {
			return nil, errors.New("no such metric")
		}
		outMt.Delta = &delta
	default:
		return nil, errors.New("no such metric")
	}

	return &outMt, nil
}

func (rs *Storage) SaveMetric(ctx context.Context, metric models.Metric) (*models.Metric, error) {
	outMt := models.Metric{ID: metric.ID, MType: metric.MType}
	var value float64
	var delta int64

	switch metric.MType {
	case "gauge":
		value = *metric.Value
		rs.memStorage.SaveGauge(metric.ID, value)
		outMt.Value = &value
	case "counter":
		delta = *metric.Delta
		rs.memStorage.SaveCounter(metric.ID, delta)
		delta = rs.memStorage.CounterMetrics[metric.ID]
		outMt.Delta = &delta
	default:
		return nil, newRAMErr("SaveMetric", errors.New("no such metric type"))
	}

	return &outMt, nil
}

func (rs *Storage) SaveBatch(ctx context.Context, metrics []models.Metric) ([]models.Metric, error) {
	var outMetrics []models.Metric
	var value float64
	var delta int64

	for _, mt := range metrics {
		outMt := models.Metric{ID: mt.ID, MType: mt.MType}

		switch mt.MType {
		case "gauge":
			value = *mt.Value
			rs.memStorage.SaveGauge(mt.ID, value)
			outMt.Value = &value
		case "counter":
			delta = *mt.Delta
			rs.memStorage.SaveCounter(mt.ID, delta)
			outMt.Delta = &delta
		default:
			return nil, newRAMErr("SaveBatch", errors.New("no such metric type"))
		}

		outMetrics = append(outMetrics, outMt)
	}

	return outMetrics, nil
}
