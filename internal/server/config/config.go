// Package config provides parsing configuration provided on application start.
// Configs can be provided through - listed accordingly precedence:
//
//		|-- flags
//		  |-- a               server address to post metric values
//		  |-- r               report interval in second to post metric values on server
//		  |-- p               metric values refreshing interval in second
//		  |-- l               limit sequential requests to server
//		  |-- k               secret key to sign request
//		  |-- crypto-key      path to public crypto key
//		  |-- c               path to config path
//		  |-- config          path to config path
//		  |-- t               trusted subnet in CIDR notation
//		|-- environment variables
//		  |-- ADDRESS         server address to post metric values
//		  |-- REPORT_INTERVAL report interval in second to post metric values on server
//		  |-- POLL_INTERVAL   metric values refreshing interval in second
//		  |-- RATE_LIMIT      limit sequential requests to server
//		  |-- KEY             secret key to sign request
//		  |-- CRYPTO_KEY      path to public crypto key
//		  |-- CONFIG          path to config path
//		  |-- TRUSTED_SUBNET  trusted subnet in CIDR notation
//		|-- config file
//		  |-- example
//		      {
//		        "address": "localhost:8080", // аналог переменной окружения ADDRESS или флага -a
//		        "report_interval": "1s", // аналог переменной окружения REPORT_INTERVAL или флага -r
//		        "poll_interval": "1s", // аналог переменной окружения POLL_INTERVAL или флага -p
//		        "crypto_key": "/path/to/key.pem", // аналог переменной окружения CRYPTO_KEY или флага -crypto-key
//	            "trusted_subnet": "127.0.0.0/31" // IP-адрес хоста агента в строковом представлении бесклассовой адресации (CIDR)
//		      }
package config

import (
	"encoding/json"
	"net"
	"os"
	"time"

	"github.com/caarlos0/env/v6"
	"go.uber.org/zap"
)

// Config struct keeps tags provided from console on application start.
type Config struct {
	Address           string `env:"ADDRESS"`
	StoreInterval     int    `env:"STORE_INTERVAL"`
	FileStoragePath   string `env:"FILE_STORAGE_PATH"`
	Restore           bool   `env:"RESTORE"`
	DatabaseDsn       string `env:"DATABASE_DSN"`
	Key               string `env:"KEY"`
	CryptoPrivKeyPath string `env:"CRYPTO_KEY"`
	ConfigPath        string `env:"CONFIG"`
	TrustedSubnetCIDR string `env:"TRUSTED_SUBNET"`
	TrustedIPNet      *net.IPNet
}

// JSONConfig stores Config options parsed from provided config file
type JSONConfig struct {
	Address           string `json:"address"`
	StoreInterval     string `json:"store_interval"`
	FileStoragePath   string `json:"store_file"`
	Restore           bool   `json:"restore"`
	DatabaseDsn       string `json:"database_dsn"`
	Key               string `json:"key"`
	CryptoPrivKeyPath string `json:"crypto_key"`
	TrustedSubnetCIDR string `json:"trusted_subnet"`
}

// Configuration default constants
const (
	storeIntervalDefault = 300
	hostDefault          = "localhost"
	portDefault          = "8080"
	fileStorageDefault   = "/tmp/metrics-db.json"
)

// Init parses configuration from flags, environment variables and file accordingly precedence
func (c *Config) Init(logger *zap.Logger) {
	// CONFIGS FROM FILE
	if c.ConfigPath != "" {
		if err := c.parseConfigFile(); err != nil {
			logger.Warn("config file reading error", zap.String("err-msg", err.Error()))
		}
	}
	// FLAGS
	if err := c.parseFlags(os.Args[1:]); err != nil {
		logger.Warn("flags parsing error", zap.String("err-msg", err.Error()))
	}

	// ENV VARS
	if err := env.Parse(c); err != nil {
		logger.Warn("parse env variables error", zap.String("err-msg", err.Error()))
	}

	if err := c.checkValues(); err != nil {
		logger.Warn("server configuration initiation error", zap.String("err-msg", err.Error()))
	}
}

// parseConfigFile parses config parameters from provided config file
func (c *Config) parseConfigFile() error {
	configBytes, err := os.ReadFile(c.ConfigPath)
	if err != nil {
		return err
	}
	var jsonConfig = JSONConfig{Restore: true}
	if err = json.Unmarshal(configBytes, &jsonConfig); err != nil {
		return err
	}

	if jsonConfig.Address != "" {
		c.Address = jsonConfig.Address
	}
	if jsonConfig.FileStoragePath != "" {
		c.FileStoragePath = jsonConfig.FileStoragePath
	}
	if jsonConfig.Restore {
		c.Restore = jsonConfig.Restore
	}
	if jsonConfig.DatabaseDsn != "" {
		c.DatabaseDsn = jsonConfig.DatabaseDsn
	}
	if jsonConfig.Key != "" {
		c.Key = jsonConfig.Key
	}
	if jsonConfig.CryptoPrivKeyPath != "" {
		c.CryptoPrivKeyPath = jsonConfig.CryptoPrivKeyPath
	}
	if jsonConfig.StoreInterval != "" {
		duration, err := time.ParseDuration(jsonConfig.StoreInterval)
		if err != nil {
			return err
		} else {
			c.StoreInterval = int(duration.Seconds())
		}
	}
	return nil
}

// setDefaults sets default values for report interval, poll interval and address if those hasn't set by flags and env vars
func (c *Config) checkValues() error {
	if c.StoreInterval <= 0 {
		c.StoreInterval = storeIntervalDefault
	}
	if c.FileStoragePath == "" {
		c.FileStoragePath = fileStorageDefault
	}
	if c.Address == "" {
		c.Address = hostDefault + ":" + portDefault
	}

	if c.TrustedSubnetCIDR != "" {
		_, trustedNet, err := net.ParseCIDR(c.TrustedSubnetCIDR)
		if err != nil {
			c.TrustedSubnetCIDR = ""
			return err
		}
		c.TrustedIPNet = trustedNet
	}

	return nil
}
