package config

import (
	"log"

	"github.com/caarlos0/env/v6"
)

func parseEnvVars() {
	err := env.Parse(&Options)
	if err != nil {
		log.Print(err)
	}

	if Options.ReportInterval <= 0 {
		Options.ReportInterval = reportIntervalDefault
	}
	if Options.PollInterval <= 0 {
		Options.PollInterval = pollIntervalDefault
	}
}
