package file

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/aykuli/observer/internal/server/logger"
	"github.com/aykuli/observer/internal/server/models"
	"github.com/aykuli/observer/internal/server/storage"
)

type Storage struct {
	memStorage    storage.MemStorage
	filePath      string
	storeInterval int
}

func NewStorage(filePath string, restore bool, storeInterval int) (*Storage, error) {
	var fs = Storage{
		memStorage: storage.MemStorage{
			GaugeMetrics:   map[string]float64{},
			CounterMetrics: map[string]int64{},
		},
		filePath:      filePath,
		storeInterval: storeInterval,
	}
	err := fs.checkFile(filePath)
	if err != nil {
		return nil, newFSError("New", err)
	}

	if restore {
		err = fs.load()
		if err != nil {
			return nil, newFSError("New", err)
		}
	}
	if storeInterval > 0 {
		go fs.startSaveMetricsTicker()
	}

	return &fs, nil
}

func (fs *Storage) checkFile(filePath string) error {
	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		if _, err = os.Create(filePath); err != nil {
			return err
		}
	}

	return nil
}

func (fs *Storage) load() error {
	consumer, err := NewConsumer(fs.filePath)
	if err != nil {
		return nil
	}

	fStore, err := consumer.ReadMetrics()
	if err != nil {
		return err
	}
	if fStore == nil {
		return nil
	}

	fs.memStorage.GaugeMetrics = fStore.GaugeMetrics
	fs.memStorage.CounterMetrics = fStore.CounterMetrics
	return nil
}

func (fs *Storage) startSaveMetricsTicker() {
	collectTicker := time.NewTicker(time.Duration(fs.storeInterval) * time.Second)
	collectQuit := make(chan struct{})
	for {
		select {
		case <-collectTicker.C:
			err := fs.saveMetricsToFile()
			if err != nil {
				logger.Log.Debug("failed metrics saving to file.", zap.Error(err))
				collectTicker.Stop()
			}
		case <-collectQuit:
			collectTicker.Stop()
		}
	}
}

func (fs *Storage) saveMetricsToFile() error {
	producer, err := NewProducer(fs.filePath)
	if err != nil {
		return err
	}
	defer producer.Close()

	if len(fs.memStorage.GaugeMetrics) > 0 || len(fs.memStorage.CounterMetrics) > 0 {
		err = producer.WriteMetrics(storage.MemStorage{
			GaugeMetrics:   fs.memStorage.GaugeMetrics,
			CounterMetrics: fs.memStorage.CounterMetrics,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (fs *Storage) Ping(ctx context.Context) error {
	return nil
}

func (fs *Storage) GetMetrics(ctx context.Context) (string, error) {
	var metrics []string
	for k, v := range fs.memStorage.GaugeMetrics {
		metrics = append(metrics, fmt.Sprintf("%s: %f", k, v))
	}
	for k, d := range fs.memStorage.CounterMetrics {
		metrics = append(metrics, fmt.Sprintf("%s: %d", k, d))
	}

	return strings.Join(metrics, ",\n"), nil
}

func (fs *Storage) ReadMetric(ctx context.Context, mName, mType string) (*models.Metric, error) {
	outMt := models.Metric{ID: mName, MType: mType}
	var value float64
	var delta int64
	switch mType {
	case "gauge":
		v, ok := fs.memStorage.GaugeMetrics[mName]
		if !ok {
			return nil, errors.New("no such metric")
		}
		value = v
		outMt.Value = &value
	case "counter":
		d, ok := fs.memStorage.CounterMetrics[mName]
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

func (fs *Storage) SaveMetric(ctx context.Context, metric models.Metric) (*models.Metric, error) {
	outMt := models.Metric{ID: metric.ID, MType: metric.MType}
	var value float64
	var delta int64

	switch metric.MType {
	case "gauge":
		value = *metric.Value
		fs.memStorage.GaugeMetrics[metric.ID] = value
		outMt.Value = &value
	case "counter":
		delta = *metric.Delta
		fs.memStorage.CounterMetrics[metric.ID] += delta
		delta = fs.memStorage.CounterMetrics[metric.ID]
		outMt.Delta = &delta
	default:
		return nil, newFSError("SaveMetric", errors.New("no such metric type"))
	}
	if fs.storeInterval == 0 {
		if err := fs.saveMetricsToFile(); err != nil {
			return nil, err
		}
	}

	return &outMt, nil
}

func (fs *Storage) SaveBatch(ctx context.Context, metrics []models.Metric) ([]models.Metric, error) {
	var outMetrics []models.Metric
	var delta int64

	for _, mt := range metrics {
		outMt := models.Metric{ID: mt.ID, MType: mt.MType}
		switch mt.MType {
		case "gauge":
			fs.memStorage.GaugeMetrics[mt.ID] = *mt.Value
		case "counter":
			delta = *mt.Delta
			fs.memStorage.CounterMetrics[mt.ID] += delta
			delta = fs.memStorage.CounterMetrics[mt.ID]
			outMt.Delta = &delta
		default:
			return nil, newFSError("SaveBatch", errors.New("no such metric type"))
		}

		outMetrics = append(outMetrics, outMt)
	}
	if fs.storeInterval == 0 {
		if err := fs.saveMetricsToFile(); err != nil {
			return nil, err
		}
	}

	return outMetrics, nil
}
