package config

import (
	"errors"
	"strings"
)

var ListenAddr = "localhost:8080"
var ReportInterval = 10
var PollInterval = 2

var MaxTries = 5

type Config struct {
	Address        string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
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

	return ServerAddr{
		Host: hp[0],
		Port: hp[1],
	}, nil
}
