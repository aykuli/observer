package config

import (
	"flag"
	"log"
	"os"
)

func parseFlags() {
	fs := flag.NewFlagSet("server", flag.ContinueOnError)
	fs.StringVar(&Options.Address, "a", hostDefault+":"+portDefault, "server address to run on")
	fs.StringVar(&Options.FileStoragePath, "f", fileStorageDefault, "path to save metrics values")
	fs.IntVar(&Options.StoreInterval, "i", 300, "metrics store interval in seconds")
	fs.BoolVar(&Options.Restore, "r", true, "restore metrics from file")
	fs.StringVar(&Options.DatabaseDsn, "d", "", "database source name")
	fs.StringVar(&Options.Key, "k", "", "secret key to sign response")

	err := fs.Parse(os.Args[1:])
	if err != nil {
		log.Print(err)
	}
}
