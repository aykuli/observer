package storage

import (
	"time"

	"go.uber.org/zap"

	"github.com/aykuli/observer/internal/errors"
	"github.com/aykuli/observer/internal/server/config"
	"github.com/aykuli/observer/internal/server/logger"
)

type GaugeMetrics map[string]float64
type CounterMetrics map[string]int64

type MemStorage struct {
	GaugeMetrics   GaugeMetrics   `json:"gauge_metrics"`
	CounterMetrics CounterMetrics `json:"counter_metrics"`
}

func (m *MemStorage) Load() error {
	consumer, err := NewConsumer(config.Options.FileStoragePath)
	if err != nil {
		return nil
	}

	mStore, err := consumer.ReadMetrics()
	if err != nil {
		return errors.NewStorageError("Load", err)
	}

	m.GaugeMetrics = mStore.GaugeMetrics
	m.CounterMetrics = mStore.CounterMetrics
	return nil
}

func (m *MemStorage) SaveMetricsPeriodically() {
	collectTicker := time.NewTicker(time.Duration(config.Options.StoreInterval) * time.Second)
	collectQuit := make(chan struct{})
	for {
		select {
		case <-collectTicker.C:
			err := m.SaveMetricsToFile()
			if err != nil {
				logger.Log.Debug("failed metrics saving to file.", zap.Error(err))
				collectTicker.Stop()
			}
		case <-collectQuit:
			collectTicker.Stop()
		}
	}
}

func (m *MemStorage) SaveMetricsToFile() error {
	producer, err := NewProducer(config.Options.FileStoragePath)
	if err != nil {
		return errors.NewStorageError("SaveMetricsToFile", err)
	}
	defer producer.Close()

	if len(m.GaugeMetrics) > 0 || len(m.CounterMetrics) > 0 {
		err = producer.WriteMetrics(MemStorage{
			GaugeMetrics:   m.GaugeMetrics,
			CounterMetrics: m.CounterMetrics,
		})
		if err != nil {
			return errors.NewStorageError("SaveMetricsToFile", err)
		}
	}

	return nil
}
