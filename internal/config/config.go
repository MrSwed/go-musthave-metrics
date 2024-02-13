package config

import (
	"flag"
	"os"
	"strconv"
	"strings"

	"github.com/MrSwed/go-musthave-metrics/internal/constant"
)

type StorageConfig struct {
	FileStoreInterval int
	FileStoragePath   string
	StorageRestore    bool
}

type Config struct {
	ServerAddress string
	DatabaseDSN   string
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

func (c *Config) Init() *Config {
	return c.withFlags().WithEnv().CleanSchemes()
}

func (c *Config) WithEnv() *Config {
	if addr, ok := os.LookupEnv(constant.EnvNameServerAddress); ok && addr != "" {
		c.ServerAddress = addr
	}
	if file, ok := os.LookupEnv(constant.EnvNameFileStoragePath); ok && file != "" {
		c.FileStoragePath = file
	}
	if sInterval, ok := os.LookupEnv(constant.EnvNameStoreInterval); ok {
		if sInterval, err := strconv.Atoi(sInterval); err == nil {
			c.FileStoreInterval = sInterval
		}
	}
	if restore, ok := os.LookupEnv(constant.EnvNameRestore); ok {
		func() {
			for _, v := range []string{"true", "1", "on", "y", "yes"} {
				if v == strings.ToLower(restore) {
					c.StorageRestore = true
					return
				}
			}
			for _, v := range []string{"false", "0", "off", "n", "no"} {
				if v == strings.ToLower(restore) {
					c.StorageRestore = false
					return
				}
			}
		}()
	}
	if dbDSN, ok := os.LookupEnv(constant.EnvNameDBDSN); ok {
		c.DatabaseDSN = dbDSN
	}
	return c
}

func (c *Config) withFlags() *Config {
	flag.StringVar(&c.ServerAddress, "a", c.ServerAddress, "Provide the address start server")
	flag.IntVar(&c.FileStoreInterval, "i", c.FileStoreInterval, "Provide the interval store (sec)")
	flag.StringVar(&c.FileStoragePath, "f", c.FileStoragePath, "Provide the file storage path")
	flag.BoolVar(&c.StorageRestore, "r", c.StorageRestore, "Restore storage at boot")
	flag.StringVar(&c.DatabaseDSN, "d", c.DatabaseDSN, "Provide the database dsn connect string")
	flag.Parse()
	return c
}

func (c *Config) CleanSchemes() *Config {
	for _, v := range []string{"http://", "https://"} {
		c.ServerAddress = strings.TrimPrefix(c.ServerAddress, v)
	}
	c.DatabaseDSN = strings.Trim(c.DatabaseDSN, "'")
	return c
}
