package main

import (
	"fmt"
	"log"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	Address string `env:"ADDRESS"`
}

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
		listenAddr = fmt.Sprintf("%s:%v", serverAddr.Host, serverAddr.Port)
	}
}
