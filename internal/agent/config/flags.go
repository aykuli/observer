package config

import "flag"

type FlagsConfig struct {
	Address          string
	ReportInterval   int
	PollInterval     int
	RateLimit        int
	Key              string
	CryptoPubKeyPath string
	ConfigPath       string
	ConfigPathLong   string
}

// parseFlags parses options from flags provided on application running
func (c *Config) parseFlags(args []string) error {
	var parsedFromFlags FlagsConfig

	fs := flag.NewFlagSet("agent", flag.ContinueOnError)
	fs.StringVar(&parsedFromFlags.Address, "a", "", "server address to post metric values")
	fs.IntVar(&parsedFromFlags.ReportInterval, "r", 0, "report interval in second to post metric values on server")
	fs.IntVar(&parsedFromFlags.PollInterval, "p", 0, "metric values refreshing interval in second")
	fs.IntVar(&parsedFromFlags.RateLimit, "l", 0, "limit sequential requests to server")
	fs.StringVar(&parsedFromFlags.Key, "k", "", "secret key to sign request")
	fs.StringVar(&parsedFromFlags.CryptoPubKeyPath, "crypto-key", "", "path to public crypto key")
	fs.StringVar(&parsedFromFlags.ConfigPath, "c", "", "path to config path")
	fs.StringVar(&parsedFromFlags.ConfigPathLong, "config", "", "path to config path")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if parsedFromFlags.Address != "" {
		c.Address = parsedFromFlags.Address
	}

	if parsedFromFlags.ReportInterval != 0 {
		c.ReportInterval = parsedFromFlags.ReportInterval
	}

	if parsedFromFlags.PollInterval != 0 {
		c.PollInterval = parsedFromFlags.PollInterval
	}

	if parsedFromFlags.RateLimit != 0 {
		c.RateLimit = parsedFromFlags.RateLimit
	}

	if parsedFromFlags.Key != "" {
		c.Key = parsedFromFlags.Key
	}

	if parsedFromFlags.CryptoPubKeyPath != "" {
		c.CryptoPubKeyPath = parsedFromFlags.CryptoPubKeyPath
	}

	if parsedFromFlags.ConfigPath != "" {
		c.ConfigPath = parsedFromFlags.ConfigPath
	}

	if parsedFromFlags.ConfigPathLong != "" {
		c.ConfigPath = parsedFromFlags.ConfigPathLong
	}

	return nil
}
