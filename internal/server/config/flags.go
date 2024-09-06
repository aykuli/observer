package config

import (
	"flag"
)

type FlagsConfig struct {
	Address           string
	StoreInterval     int
	FileStoragePath   string
	Restore           bool
	DatabaseDsn       string
	Key               string
	CryptoPrivKeyPath string
	ConfigPath        string
	ConfigPathLong    string
}

// parseFlags parses options from flags provided on application running
func (c *Config) parseFlags(args []string) error {
	var parsedFromFlags FlagsConfig
	fs := flag.NewFlagSet("server", flag.ContinueOnError)
	fs.StringVar(&parsedFromFlags.Address, "a", "", "server address to run on")
	fs.StringVar(&parsedFromFlags.FileStoragePath, "f", "", "path to save metrics values")
	fs.IntVar(&parsedFromFlags.StoreInterval, "i", 0, "metrics store interval in seconds")
	fs.BoolVar(&parsedFromFlags.Restore, "r", true, "restore metrics from file")
	fs.StringVar(&parsedFromFlags.DatabaseDsn, "d", "", "database source name")
	fs.StringVar(&parsedFromFlags.Key, "k", "", "secret key to sign response")
	fs.StringVar(&parsedFromFlags.CryptoPrivKeyPath, "crypto-key", "", "path to private crypto key")
	fs.StringVar(&parsedFromFlags.ConfigPath, "c", "", "path to config file")
	fs.StringVar(&parsedFromFlags.ConfigPathLong, "config", "", "path to config file")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if parsedFromFlags.Address != "" {
		c.Address = parsedFromFlags.Address
	}

	if parsedFromFlags.FileStoragePath != "" {
		c.FileStoragePath = parsedFromFlags.FileStoragePath
	}

	if parsedFromFlags.StoreInterval != 0 {
		c.StoreInterval = parsedFromFlags.StoreInterval
	}

	if !parsedFromFlags.Restore {
		c.Restore = parsedFromFlags.Restore
	}

	if parsedFromFlags.DatabaseDsn != "" {
		c.DatabaseDsn = parsedFromFlags.DatabaseDsn
	}

	if parsedFromFlags.Key != "" {
		c.Key = parsedFromFlags.Key
	}

	if parsedFromFlags.CryptoPrivKeyPath != "" {
		c.CryptoPrivKeyPath = parsedFromFlags.CryptoPrivKeyPath
	}

	if parsedFromFlags.ConfigPath != "" {
		c.ConfigPath = parsedFromFlags.ConfigPath
	}

	if parsedFromFlags.ConfigPathLong != "" {
		c.ConfigPath = parsedFromFlags.ConfigPathLong
	}

	return nil
}
