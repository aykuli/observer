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

	if serverAddrStr := Options.Address; serverAddrStr != "" {
		serverAddr, err := parseServerAddr(serverAddrStr)
		if err != nil {
			log.Print(err)
		}

		Options.Address = serverAddr.String()
	}

	if filename := Options.FileStoragePath; filename != "" {
		writeFile, err := needWriteFile(filename)
		if err != nil {
			log.Print(err)
		}

		Options.SaveMetrics = writeFile
	}

	if Options.StoreInterval < 0 {
		Options.StoreInterval = storeIntervalDefault
	}
}
