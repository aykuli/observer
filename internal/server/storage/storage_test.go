package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemStorage_SaveMetricsToFile(t *testing.T) {
	emptyMemStorage := MemStorage{
		GaugeMetrics:   GaugeMetrics{},
		CounterMetrics: CounterMetrics{},
	}

	existsMemStorage := MemStorage{
		GaugeMetrics:   GaugeMetrics{"Alloc": 2298632, "BuckHashSys": 7290},
		CounterMetrics: CounterMetrics{"PollCount": 256},
	}

	t.Run("saves metrics values into file, then load it to empty memStroage", func(t *testing.T) {
		err := existsMemStorage.SaveMetricsToFile()
		require.NoError(t, err)

		err = emptyMemStorage.Load()
		require.NoError(t, err)

		assert.Equal(t, existsMemStorage, emptyMemStorage, "should be equal")

		require.NoError(t, err)
	})
}
