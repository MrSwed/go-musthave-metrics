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
	StorageConfig
}

func NewConfig(init ...bool) *Config {
	c := &Config{
		ServerAddress: serverAddress,
		StorageConfig: StorageConfig{
			StoreInterval:   storeInterval,
			FileStoragePath: fileStoragePath,
			StorageRestore:  storageRestore,
		},
	}
	if len(init) > 0 && init[0] {
		return c.withFlags().withEnv().cleanSchemes()
	}
	return c
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
	return c
}

func (c *Config) withFlags() *Config {
	flag.StringVar(&c.ServerAddress, "a", serverAddress, "Provide the address start server")
	flag.IntVar(&c.StoreInterval, "i", storeInterval, "Provide the interval store (sec)")
	flag.StringVar(&c.FileStoragePath, "f", fileStoragePath, "Provide the file storage path")
	flag.BoolVar(&c.StorageRestore, "r", storageRestore, "Restore storage at boot")
	flag.Parse()
	return c
}

func (c *Config) cleanSchemes() *Config {
	c.ServerAddress = strings.TrimPrefix(c.ServerAddress, "http://")
	c.ServerAddress = strings.TrimPrefix(c.ServerAddress, "https://")
	return c
}
