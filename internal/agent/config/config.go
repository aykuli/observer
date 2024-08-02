// Package config provides parsing configuration provided on application start.
package config

import "os"

// Config struct keeps tags provided from console on application start.
type Config struct {
	Address        string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	Key            string `env:"KEY"`
	RateLimit      int    `env:"RATE_LIMIT"`
}

// Configuration default constants
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

// ServerAddr struct provides server host and port.
type ServerAddr struct {
	Host string
	Port string
}

func init() {
	parseFlags(os.Args[1:])
	parseEnvVars()
}
