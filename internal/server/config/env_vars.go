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

	if Options.StoreInterval < 0 {
		Options.StoreInterval = storeIntervalDefault
	}
}
