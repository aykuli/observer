package config

import (
	"log"

	"github.com/caarlos0/env/v6"
)

func parseEnvVars() {
	var envVars Config
	err := env.Parse(&envVars)
	if err != nil {
		log.Fatal(err)
	}

	if serverAddrStr := envVars.Address; serverAddrStr != "" {
		serverAddr, err := parseServerAddr(serverAddrStr)
		if err != nil {
			log.Fatal(err)
		}
		ListenAddr = serverAddr.Host + ":" + serverAddr.Port
	}
}
