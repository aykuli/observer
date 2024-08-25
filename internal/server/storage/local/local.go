// Package local provides handler to save metrics in local file.
package local

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/aykuli/observer/internal/models"
	"github.com/aykuli/observer/internal/server/config"
	"github.com/aykuli/observer/internal/server/storage"
)

type Storage struct {
	memStorage storage.MetricsMap
	logger     zap.SugaredLogger
}

func NewStorage(options config.Config, logger zap.SugaredLogger) (*Storage, error) {
	flushOnSave := options.FileStoragePath != "" && options.StoreInterval == 0
	s := Storage{
		memStorage: *storage.NewMetricsMap(options.FileStoragePath, flushOnSave),
		logger:     logger,
	}

	if options.FileStoragePath != "" {
		err := s.checkFile(options.FileStoragePath)
		if err != nil {
			return nil, newFSError("New", err)
		}

		if options.Restore {
			if err = s.memStorage.LoadFromFile(); err != nil {
				return nil, newFSError("New", err)
			}
		}

		if options.StoreInterval > 0 {
			go s.startSaveMetricsTicker(options.StoreInterval)
		}
	}

	return &s, nil
}

func (s *Storage) checkFile(filePath string) error {
	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		if _, err = os.Create(filePath); err != nil {
			return err
		}
	}

	return nil
}

func (s *Storage) startSaveMetricsTicker(storeInterval int) {
	collectTicker := time.NewTicker(time.Duration(storeInterval) * time.Second)
	defer collectTicker.Stop()
	for range collectTicker.C {
		err := s.memStorage.SaveToFile()
		if err != nil {
			s.logger.Errorln("failed metrics saving to local.", zap.Error(err))
		}
	}
}

func (s *Storage) Ping(ctx context.Context) error {
	return nil
}

func (s *Storage) GetMetrics(ctx context.Context) (string, error) {
	gaugeMts := s.memStorage.GetGaugeMetrics()
	counterMts := s.memStorage.GetCounterMetrics()
	var metrics = make([]string, len(gaugeMts)+len(counterMts))
	i := 0
	for k, v := range gaugeMts {
		metrics[i] = fmt.Sprintf("%s: %f", k, v)
		i++
	}
	for k, d := range counterMts {
		metrics[i] = fmt.Sprintf("%s: %d", k, d)
		i++
	}

	sort.Slice(metrics, func(i, j int) bool { return metrics[i] < metrics[j] })

	return strings.Join(metrics, ",\n"), nil
}

func (s *Storage) ReadMetric(ctx context.Context, mName, mType string) (*models.Metric, error) {
	outMt := models.Metric{ID: mName, MType: mType}
	var value float64
	var delta int64

	switch mType {
	case "gauge":
		v, ok := s.memStorage.GetGauge(mName)
		if !ok {
			return nil, errors.New("no such metric")
		}
		value = v
		outMt.Value = &value
	case "counter":
		d, ok := s.memStorage.GetCounter(mName)
		if !ok {
			return nil, errors.New("no such metric")
		}
		delta = d
		outMt.Delta = &delta
	default:
		return nil, newFSError("ReadMetric", errors.New("no such metric"))
	}

	return &outMt, nil
}

func (s *Storage) SaveMetric(ctx context.Context, metric models.Metric) (*models.Metric, error) {
	outMt := models.Metric{ID: metric.ID, MType: metric.MType}
	var value float64
	var delta int64

	switch metric.MType {
	case "gauge":
		value = *metric.Value
		newValue, err := s.memStorage.SaveGauge(metric.ID, value)
		if err != nil {
			return nil, newFSError("SaveMetric", err)
		}
		outMt.Value = &newValue
	case "counter":
		delta = *metric.Delta
		newDelta, err := s.memStorage.SaveCounter(metric.ID, delta)
		if err != nil {
			return nil, newFSError("SaveMetric", err)
		}
		outMt.Delta = &newDelta
	default:
		return nil, newFSError("SaveMetric", errors.New("no such metric type"))
	}

	return &outMt, nil
}

func (s *Storage) SaveBatch(ctx context.Context, metrics []models.Metric) ([]models.Metric, error) {
	var outMetrics = make([]models.Metric, len(metrics))
	var delta int64

	for i, mt := range metrics {
		outMt := models.Metric{ID: mt.ID, MType: mt.MType}
		switch mt.MType {
		case "gauge":
			newValue, err := s.memStorage.SaveGauge(mt.ID, *mt.Value)
			if err != nil {
				return nil, newFSError("SaveBatch", err)
			}
			outMt.Value = &newValue
		case "counter":
			delta = *mt.Delta
			newDelta, err := s.memStorage.SaveCounter(mt.ID, delta)
			if err != nil {
				return nil, newFSError("SaveBatch", err)
			}
			outMt.Delta = &newDelta
		default:
			return nil, newFSError("SaveBatch", errors.New("no such metric type"))
		}

		outMetrics[i] = outMt
	}

	return outMetrics, nil
}
