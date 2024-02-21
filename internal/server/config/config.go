package config

import (
	"errors"
	"strings"
)

var ListenAddr = "localhost:8080"

type Config struct {
	Address string `env:"ADDRESS"`
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
