package config

import (
	"flag"
	"os"
	"strconv"
	"strings"

	"github.com/MrSwed/go-musthave-metrics/internal/server/constant"
)

// StorageConfig file storage configs
type StorageConfig struct {
	FileStoragePath   string
	StorageRestore    bool
	FileStoreInterval int
}

// WEB  config
type WEB struct {
	Key string
}

// Config all configs
type Config struct {
	ServerAddress string
	DatabaseDSN   string
	WEB
	StorageConfig
}

func NewConfig() *Config {
	return &Config{
		ServerAddress: constant.ServerAddress,
		StorageConfig: StorageConfig{
			FileStoreInterval: constant.StoreInterval,
			FileStoragePath:   constant.FileStoragePath,
			StorageRestore:    constant.StorageRestore,
		},
	}
}

// Init all configs
func (c *Config) Init() *Config {
	return c.withFlags().WithEnv().CleanSchemes()
}

// WithEnv gets ENV configs
func (c *Config) WithEnv() *Config {
	if envVal, ok := os.LookupEnv(constant.EnvNameServerAddress); ok && envVal != "" {
		c.ServerAddress = envVal
	}
	if envVal, ok := os.LookupEnv(constant.EnvNameFileStoragePath); ok && envVal != "" {
		c.FileStoragePath = envVal
	}
	if envVal, ok := os.LookupEnv(constant.EnvNameStoreInterval); ok {
		if sInterval, err := strconv.Atoi(envVal); err == nil {
			c.FileStoreInterval = sInterval
		}
	}
	if envVal, ok := os.LookupEnv(constant.EnvNameRestore); ok {
		func() {
			for _, v := range []string{"true", "1", "on", "y", "yes"} {
				if v == strings.ToLower(envVal) {
					c.StorageRestore = true
					return
				}
			}
			for _, v := range []string{"false", "0", "off", "n", "no"} {
				if v == strings.ToLower(envVal) {
					c.StorageRestore = false
					return
				}
			}
		}()
	}
	if envVal, ok := os.LookupEnv(constant.EnvNameDBDSN); ok {
		c.DatabaseDSN = envVal
	}
	if envVal, ok := os.LookupEnv(constant.EnvNameKey); ok {
		c.Key = envVal
	}

	return c
}

func (c *Config) withFlags() *Config {
	flag.StringVar(&c.ServerAddress, "a", c.ServerAddress, "Provide the address start server")
	flag.IntVar(&c.FileStoreInterval, "i", c.FileStoreInterval, "Provide the interval store (sec)")
	flag.StringVar(&c.FileStoragePath, "f", c.FileStoragePath, "Provide the file storage path")
	flag.BoolVar(&c.StorageRestore, "r", c.StorageRestore, "Restore storage at boot")
	flag.StringVar(&c.DatabaseDSN, "d", c.DatabaseDSN, "Provide the database dsn connect string")
	flag.StringVar(&c.Key, "k", "", "Provide the key")
	flag.Parse()
	return c
}

// CleanSchemes check and repair config parameters
func (c *Config) CleanSchemes() *Config {
	for _, v := range []string{"http://", "https://"} {
		c.ServerAddress = strings.TrimPrefix(c.ServerAddress, v)
	}
	c.DatabaseDSN = strings.Trim(c.DatabaseDSN, "'")
	return c
}
