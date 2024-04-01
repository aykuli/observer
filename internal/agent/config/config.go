package config

type Config struct {
	Address        string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
}

const (
	reportIntervalDefault = 10
	pollIntervalDefault   = 10
	hostDefault           = "localhost"
	portDefault           = "8080"
)

var Options = Config{
	Address:        hostDefault + ":" + portDefault,
	ReportInterval: reportIntervalDefault,
	PollInterval:   pollIntervalDefault,
}

type ServerAddr struct {
	Host string
	Port string
}

func init() {
	parseFlags()
	parseEnvVars()
}
