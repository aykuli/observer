package storage

import (
	"encoding/json"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aykuli/observer/internal/server/config"
)

const (
	testJSONFileName = "test.json"
)

func TestMemStorage_Load(t *testing.T) {
	emptyMemStorage := MemStorage{
		GaugeMetrics:   GaugeMetrics{},
		CounterMetrics: CounterMetrics{},
	}

	existsMemStorage := MemStorage{
		GaugeMetrics:   GaugeMetrics{"Alloc": 2298632, "BuckHashSys": 7290},
		CounterMetrics: CounterMetrics{"PollCount": 256},
	}

	t.Run("loads memstorage from file", func(t *testing.T) {
		modeVal, err := strconv.ParseUint("0777", 8, 32)
		require.NoError(t, err)
		err = os.WriteFile(config.Options.Path, []byte(testJSONFileName), os.FileMode(modeVal))
		require.NoError(t, err)

		data, err := json.MarshalIndent(existsMemStorage, "", "  ")
		require.NoError(t, err)

		err = os.WriteFile(testJSONFileName, data, os.FileMode(modeVal))
		require.NoError(t, err)

		err = emptyMemStorage.Load()
		require.NoError(t, err)

		assert.Equal(t, emptyMemStorage, existsMemStorage, "should be loaded metrics from file")

		err = os.Remove(config.Options.Path)
		require.NoError(t, err)
		err = os.Remove(testJSONFileName)
		require.NoError(t, err)
	})
}

func TestMemStorage_SaveMetricsToFile(t *testing.T) {
	emptyMemStorage := MemStorage{
		GaugeMetrics:   GaugeMetrics{},
		CounterMetrics: CounterMetrics{},
	}

	existsMemStorage := MemStorage{
		GaugeMetrics:   GaugeMetrics{"Alloc": 2298632, "BuckHashSys": 7290},
		CounterMetrics: CounterMetrics{"PollCount": 256},
	}

	t.Run("saves metrics values into file", func(t *testing.T) {
		config.Options.FileStoragePath = "tmp.json"

		err := existsMemStorage.SaveMetricsToFile()
		require.NoError(t, err)

		data, err := os.ReadFile(config.Options.FileStoragePath)
		require.NoError(t, err)

		err = json.Unmarshal(data, &emptyMemStorage)
		require.NoError(t, err)

		assert.Equal(t, emptyMemStorage, existsMemStorage, "should be equal")

		err = os.Remove(config.Options.FileStoragePath)
		require.NoError(t, err)
	})
}
