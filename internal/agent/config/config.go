package config

import (
	"errors"
	"strings"
)

type Config struct {
	Address        string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	MaxTries       int
}

const (
	reportIntervalDefault = 10
	pollIntervalDefault   = 10
	hostDefault           = "localhost"
	portDefault           = "8080"
	schemeDefault         = "http://"
)

var Options = Config{
	Address:        schemeDefault + hostDefault + portDefault,
	ReportInterval: reportIntervalDefault,
	PollInterval:   pollIntervalDefault,
	MaxTries:       5,
}

type ServerAddr struct {
	Host string
	Port string
}

func init() {
	parseFlags()
	parseEnvVars()
}
func parseServerAddr(s string) (ServerAddr, error) {
	hp := strings.Split(s, ":")

	if len(hp) != 2 {
		return ServerAddr{}, errors.New("need address in a form host:port")
	}

	if hp[0] == "" {
		hp[0] = hostDefault
	}
	if hp[1] == "" {
		hp[1] = portDefault
	}

	return ServerAddr{
		Host: hp[0],
		Port: hp[1],
	}, nil
}
