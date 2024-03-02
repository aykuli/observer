package config

type Config struct {
	Address         string `env:"ADDRESS"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
	DatabaseDsn     string `env:"DATABASE_DSN"`
}

const (
	storeIntervalDefault = 300
	hostDefault          = "localhost"
	portDefault          = "8080"
	fileStorageDefault   = "/tmp/metrics-db.json"
)

var Options = Config{
	Address:         hostDefault + ":" + portDefault,
	StoreInterval:   storeIntervalDefault,
	FileStoragePath: fileStorageDefault,
	Restore:         true,
	DatabaseDsn:     "",
}

func init() {
	parseFlags()
	parseEnvVars()
}
