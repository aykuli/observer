package config

type Config struct {
	Path            string
	Address         string `env:"ADDRESS"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	SaveMetrics     bool
	Restore         bool `env:"RESTORE"`
}

const (
	storeIntervalDefault = 300
	hostDefault          = "localhost"
	portDefault          = "8080"
	fileStorageDefault   = "/tmp/metrics-db.json"
)

var Options = Config{
	Path:            ".server.rc",
	Address:         hostDefault + portDefault,
	StoreInterval:   storeIntervalDefault,
	FileStoragePath: fileStorageDefault,
	SaveMetrics:     true,
	Restore:         true,
}

func init() {
	parseFlags()
	parseEnvVars()
}
