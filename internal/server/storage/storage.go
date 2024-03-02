package storage

import (
	"context"

	"github.com/aykuli/observer/internal/server/models"
)

type GaugeMetrics map[string]float64
type CounterMetrics map[string]int64

type MemStorage struct {
	GaugeMetrics   GaugeMetrics   `json:"gauge_metrics"`
	CounterMetrics CounterMetrics `json:"counter_metrics"`
}

type Storage interface {
	Ping(ctx context.Context) error
	GetMetrics(ctx context.Context) (string, error)
	ReadMetric(ctx context.Context, metricName, metricType string) (*models.Metric, error)
	SaveMetric(ctx context.Context, metric models.Metric) (*models.Metric, error)
	SaveBatch(ctx context.Context, metrics []models.Metric) ([]models.Metric, error)
}
