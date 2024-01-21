package config

import (
	"flag"
	"os"
	"strings"
)

type Config struct {
	ServerAddress string
}

func NewConfig(init ...bool) *Config {
	c := &Config{ServerAddress: ServerAddress}
	if len(init) > 0 && init[0] {
		return c.withFlags().withEnv().cleanSchemes()
	}
	return c
}

func (c *Config) withEnv() *Config {
	serverAddress := os.Getenv("SERVER_ADDRESS")
	if serverAddress != "" {
		c.ServerAddress = serverAddress
	}
	return c
}

func (c *Config) withFlags() *Config {
	flag.StringVar(&c.ServerAddress, "a", c.ServerAddress, "Provide the address start server")
	flag.Parse()
	return c
}

func (c *Config) cleanSchemes() *Config {
	c.ServerAddress = strings.TrimPrefix(c.ServerAddress, "http://")
	c.ServerAddress = strings.TrimPrefix(c.ServerAddress, "https://")
	return c
}
