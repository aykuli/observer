package repository

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"

	"github.com/aykuli/observer/internal/models"
	"github.com/aykuli/observer/internal/server/config"
)

var envFolder = "../../../.env"

func TestMetricsRepository(t *testing.T) {
	ctx := context.Background()
	dsn := config.Options.DatabaseDsn
	if dsn == "" {
		envFile, err := godotenv.Read(envFolder)
		fmt.Println(envFile)
		require.NoError(t, err)
		dsn = envFile["POSTGRES_TEST_DSN"]
	}
	if dsn == "" {
		return
	}
	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)

	conn, err := pool.Acquire(ctx)
	require.NoError(t, err)
	require.NotNil(t, conn)

	repository := NewMetricsRepository(conn)

	err = repository.InitTable(ctx)
	require.NoError(t, err)

	value := rand.Float64()
	metric := models.Metric{
		ID:    "test_repository",
		MType: "gauge",
		Delta: nil,
		Value: &value,
	}
	tx, err := conn.Begin(ctx)
	require.NoError(t, err)
	savedMetric, err := repository.Save(ctx, tx, metric)
	require.NoError(t, err)
	err = tx.Commit(ctx)
	require.NoError(t, err)

	require.Equal(t, metric.ID, savedMetric.ID)
	require.Equal(t, *metric.Value, *savedMetric.Value)

	values := []float64{rand.Float64(), rand.Float64()}
	metricsBatch := []models.Metric{
		{ID: "test_1", MType: "gauge", Value: &values[0]},
		{ID: "test_2", MType: "gauge", Value: &values[1]},
	}

	tx, err = conn.Begin(ctx)
	require.NoError(t, err)
	savedMetrics, err := repository.SaveBatch(ctx, tx, metricsBatch)
	require.NoError(t, err)
	err = tx.Commit(ctx)
	require.NoError(t, err)

	require.Contains(t, savedMetrics, metricsBatch[0])
	require.Contains(t, savedMetrics, metricsBatch[1])

	metrics, err := repository.SelectAllValues(ctx)
	require.NoError(t, err)
	require.Contains(t, metrics, metric)
	require.Contains(t, metrics, metricsBatch[0])
	require.Contains(t, metrics, metricsBatch[1])
}
