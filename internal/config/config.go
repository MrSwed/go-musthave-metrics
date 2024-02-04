package config

import (
	"flag"
	"os"
	"strconv"
	"strings"
)

type StorageConfig struct {
	StoreInterval   int
	FileStoragePath string
	StorageRestore  bool
}

type Config struct {
	ServerAddress string
	DatabaseDSN   string
	StorageConfig
}

func NewConfig() *Config {
	return &Config{
		ServerAddress: serverAddress,
		StorageConfig: StorageConfig{
			StoreInterval:   storeInterval,
			FileStoragePath: fileStoragePath,
			StorageRestore:  storageRestore,
		},
	}
}

func (c *Config) Init() *Config {
	return c.withFlags().withEnv().cleanSchemes()
}

func (c *Config) withEnv() *Config {
	if addr, ok := os.LookupEnv(envNameServerAddress); ok && addr != "" {
		c.ServerAddress = addr
	}
	if file, ok := os.LookupEnv(envNameFileStoragePath); ok && file != "" {
		c.FileStoragePath = file
	}
	if sInterval, ok := os.LookupEnv(envNameStoreInterval); ok {
		if sInterval, err := strconv.Atoi(sInterval); err == nil {
			c.StoreInterval = sInterval
		}
	}
	if restore, ok := os.LookupEnv(envNameRestore); ok {
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
	if dbDSN, ok := os.LookupEnv(envNameDBDSN); ok {
		c.DatabaseDSN = dbDSN
	}
	return c
}

func (c *Config) withFlags() *Config {
	flag.StringVar(&c.ServerAddress, "a", c.ServerAddress, "Provide the address start server")
	flag.IntVar(&c.StoreInterval, "i", c.StoreInterval, "Provide the interval store (sec)")
	flag.StringVar(&c.FileStoragePath, "f", c.FileStoragePath, "Provide the file storage path")
	flag.BoolVar(&c.StorageRestore, "r", c.StorageRestore, "Restore storage at boot")
	flag.StringVar(&c.DatabaseDSN, "d", c.DatabaseDSN, "Provide the database dsn connect string")
	flag.Parse()
	return c
}

func (c *Config) cleanSchemes() *Config {
	for _, v := range []string{"http://", "https://"} {
		c.ServerAddress = strings.TrimPrefix(c.ServerAddress, v)
	}
	return c
}
