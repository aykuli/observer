package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestParseConfigFile(t *testing.T) {
	logger := zap.NewExample()
	defer logger.Sync()

	t.Run("Default configuration", func(t *testing.T) {
		var c Config
		c.Init(logger)
		require.Equal(t, c.Address, hostDefault+":"+portDefault)
		require.Equal(t, c.ReportInterval, reportIntervalDefault)
		require.Equal(t, c.PollInterval, pollIntervalDefault)
	})

	t.Run("Set configuration", func(t *testing.T) {
		configFName := "test.json"
		testConfig := `{
  "address": "testhost:5678",
  "report_interval": "1s",
  "poll_interval": "8s",
  "crypto_key": "/path/to/key.pem"
}
`
		err := os.WriteFile(configFName, []byte(testConfig), 06640)
		require.NoError(t, err)

		var c = Config{ConfigPath: configFName}
		c.Init(logger)
		require.Equal(t, c.Address, "testhost:5678")
		require.Equal(t, c.ReportInterval, 1)
		require.Equal(t, c.PollInterval, 8)
		require.Equal(t, c.RateLimit, 0)
		require.Equal(t, c.Key, "")
		require.Equal(t, c.CryptoPubKeyPath, "/path/to/key.pem")

		err = os.Remove(configFName)
		require.NoError(t, err)
		require.NoFileExists(t, configFName)
	})
}

func Example() {
	logger := zap.NewExample()
	defer logger.Sync()

	address := ":5050"
	reportInterval := "66"
	pollInterval := "77"
	key := "aynur_is_beautiful_name_fit_to_be_key"
	rateLimit := "88"
	testFlags := map[string]string{
		"a": address,
		"r": reportInterval,
		"p": pollInterval,
		"k": key,
		"l": rateLimit}
	arr := make([]string, 6)
	i := 0
	for k, v := range testFlags {
		arr[i] = "-" + k + "=" + v
		i++
	}

	var c = Config{}
	err := c.parseFlags(arr)
	fmt.Println("err:", err)

	fmt.Println(c.Address)
	fmt.Println(c.ReportInterval)
	fmt.Println(c.PollInterval)
	fmt.Println(c.Key)
	fmt.Println(c.RateLimit)

	//	Output:
	// err: <nil>
	// :5050
	// 66
	// 77
	// aynur_is_beautiful_name_fit_to_be_key
	// 88
}
