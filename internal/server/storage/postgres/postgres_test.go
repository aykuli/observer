package postgres

import (
	"context"
	"math/rand"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/aykuli/observer/internal/models"
	"github.com/aykuli/observer/internal/server/config"
)

var envFolder = "../../../../.env"

func TestPostgresStorage(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewExample()
	defer logger.Sync()
	var c config.Config
	c.Init(logger)

	if c.DatabaseDsn == "" {
		envFile, err := godotenv.Read(envFolder)
		require.NoError(t, err)
		c.DatabaseDsn = envFile["POSTGRES_TEST_DSN"]
	}
	if c.DatabaseDsn == "" {
		return
	}

	// test NewStorage
	dbStorage, err := NewStorage(c.DatabaseDsn)
	require.NoError(t, err)
	require.NotNil(t, dbStorage)

	err = dbStorage.Ping(ctx)
	require.NoError(t, err)

	// test SaveMetric
	value := rand.Float64()
	metric := models.Metric{
		ID:    "test_postgres",
		MType: "gauge",
		Delta: nil,
		Value: &value,
	}
	outMetric, err := dbStorage.SaveMetric(ctx, metric)
	require.NoError(t, err)
	require.Equal(t, *outMetric, metric)

	// test ReadMetric
	readMetric, err := dbStorage.ReadMetric(ctx, metric.ID, "gauge")
	require.NoError(t, err)
	require.Equal(t, *readMetric, metric)

	// test SaveBatch
	values := []float64{rand.Float64(), rand.Float64()}
	metricsBatch := []models.Metric{
		{ID: "test_1", MType: "gauge", Value: &values[0]},
		{ID: "test_2", MType: "gauge", Value: &values[1]},
	}
	outMetrics, err := dbStorage.SaveBatch(ctx, metricsBatch)
	require.NoError(t, err)
	require.Contains(t, outMetrics, metricsBatch[0])
	require.Contains(t, outMetrics, metricsBatch[1])
}
