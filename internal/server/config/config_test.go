package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestInit(t *testing.T) {
	logger := zap.NewExample()
	defer logger.Sync()
	t.Run("Default configuration", func(t *testing.T) {
		var c Config
		c.Init(logger)
		require.Equal(t, c.Address, hostDefault+":"+portDefault)
		require.Equal(t, c.StoreInterval, storeIntervalDefault)
		require.Equal(t, c.FileStoragePath, fileStorageDefault)
	})

	t.Run("Set configuration", func(t *testing.T) {
		configFName := "test.json"
		testConfig := `{
  "address": "some-address:1234",
  "restore": true, 
  "store_interval": "1s",
  "store_file": "/path/to/file.db",
  "database_dsn": "some-database-dsn", 
  "crypto_key": "/path/to/key.pem"
}
`
		err := os.WriteFile(configFName, []byte(testConfig), 06640)
		require.NoError(t, err)

		var c = Config{ConfigPath: configFName}
		c.Init(logger)
		require.Equal(t, c.Address, "some-address:1234")
		require.Equal(t, c.Restore, true)
		require.Equal(t, c.StoreInterval, 1)
		require.Equal(t, c.FileStoragePath, "/path/to/file.db")
		require.Equal(t, c.DatabaseDsn, "some-database-dsn")
		require.Equal(t, c.CryptoPrivKeyPath, "/path/to/key.pem")

		err = os.Remove(configFName)
		require.NoError(t, err)
		require.NoFileExists(t, configFName)
	})
}

func Example() {
	var c = Config{}

	address := ":5050"
	fileStoragePath := "awesomePath.txt"
	storeInterval := "1"
	restore := "false"
	databaseDsn := "some_dsn_path"
	key := "aynur_is_beautiful_name_fit_to_be_key"
	testFlags := map[string]string{
		"a": address,
		"f": fileStoragePath,
		"i": storeInterval,
		"r": restore,
		"d": databaseDsn,
		"k": key}
	arr := make([]string, 6)
	i := 0
	for k, v := range testFlags {
		arr[i] = "-" + k + "=" + v
		i++
	}

	err := c.parseFlags(arr)
	fmt.Println("err:", err)
	fmt.Println("Address:", c.Address)
	fmt.Println("FileStoragePath:", c.FileStoragePath)
	fmt.Println("StoreInterval:", c.StoreInterval)
	fmt.Println("Restore:", c.Restore)
	fmt.Println("DatabaseDsn:", c.DatabaseDsn)
	fmt.Println("Key:", c.Key)

	// Output:
	// err: <nil>
	// Address: :5050
	// FileStoragePath: awesomePath.txt
	// StoreInterval: 1
	// Restore: false
	// DatabaseDsn: some_dsn_path
	// Key: aynur_is_beautiful_name_fit_to_be_key
}
