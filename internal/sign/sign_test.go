package sign

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aykuli/observer/internal/models"
)

func TestSign(t *testing.T) {
	value := 5.896
	metric := models.Metric{
		ID:    "random metric",
		MType: "gauge",
		Delta: nil,
		Value: &value,
	}
	key := "my awesome secret key"
	t.Run("sign verify with right key should return true", func(t *testing.T) {
		byteData, err := json.Marshal(metric)
		require.NoError(t, err)
		hashString := GetHmacString(byteData, key)

		equal := Verify(metric, key, hashString)
		assert.True(t, equal)
	})

	t.Run("sign verify should return false if key was wrong", func(t *testing.T) {
		byteData, err := json.Marshal(metric)
		require.NoError(t, err)
		hashString := GetHmacString(byteData, key)

		equal := Verify(metric, "another key", hashString)
		assert.False(t, equal)
	})
}
