package local

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/aykuli/observer/internal/models"
	"github.com/aykuli/observer/internal/server/config"
	"github.com/aykuli/observer/internal/server/logger"
	"github.com/aykuli/observer/internal/server/storage"
)

type Storage struct {
	memStorage storage.MemStorage
}

func NewStorage(options config.Config) (*Storage, error) {
	flushOnSave := options.FileStoragePath != "" && options.StoreInterval == 0
	s := Storage{memStorage: storage.NewMemStorage(options.FileStoragePath, flushOnSave)}

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
			logger.Log.Debug("failed metrics saving to local.", zap.Error(err))
		}

	}
}

func (s *Storage) Close() {}

func (s *Storage) Ping(ctx context.Context) error {
	return nil
}

func (s *Storage) GetMetrics(ctx context.Context) (string, error) {
	var metrics []string
	for k, v := range s.memStorage.GetGaugeMetrics() {
		metrics = append(metrics, fmt.Sprintf("%s: %f", k, v))
	}
	for k, d := range s.memStorage.GetCounterMetrics() {
		metrics = append(metrics, fmt.Sprintf("%s: %d", k, d))
	}

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
	var outMetrics []models.Metric
	var delta int64

	for _, mt := range metrics {
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

		outMetrics = append(outMetrics, outMt)
	}

	return outMetrics, nil
}
