package main

import (
	"fmt"
	"log"
	"time"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	Address        string        `env:"ADDRESS"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
	PollInterval   time.Duration `env:"POLL_INTERVAL"`
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

	if envVars.ReportInterval != time.Duration(0) {
		reportInterval = envVars.ReportInterval
	}
	if envVars.PollInterval != time.Duration(0) {
		pollInterval = envVars.PollInterval
	}
}
