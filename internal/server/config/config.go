package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"flag"
	"os"
	"strconv"
	"strings"

	"github.com/MrSwed/go-musthave-metrics/internal/server/constant"

	"github.com/caarlos0/env/v10"
	"github.com/ucarion/structflag"
)

// StorageConfig file storage configs
type StorageConfig struct {
	FileStoragePath   string `env:"FILE_STORAGE_PATH" json:"file_storage_path" flag:"f" usage:"Provide the file storage path"`
	StorageRestore    bool   `env:"RESTORE" json:"restore" flag:"r" usage:"Provide the file storage path"`
	FileStoreInterval int    `env:"FILE_STORE_INTERVAL" json:"file_store_interval" flag:"i" usage:"Provide the interval in seconds"`
}

// WEB  config
type WEB struct {
	cryptoKey *rsa.PrivateKey
	Key       string `env:"KEY" json:"key" flag:"k" usage:"Private theKey"`
	CryptoKey string `env:"CRYPTO_KEY"  json:"crypto_key" flag:"crypto-key" usage:"Provide the private server key for decryption"`
}

// Config all configs
type Config struct {
	Address     string `env:"ADDRESS" json:"address"  flag:"a" usage:"Provide the address start server"`
	DatabaseDSN string `env:"DATABASE_DSN" json:"database_dsn" flag:"d" usage:"Provide the database dsn connect string"`
	Config      string `json:"-" env:"CONFIG" flag:"config" usage:"Provide file with config"`
	Config2     string `json:"-" env:"-" flag:"c" usage:"same as -config"` // ?
	WEB
	StorageConfig
}

func NewConfig() *Config {
	return &Config{
		Address: constant.ServerAddress,
		StorageConfig: StorageConfig{
			FileStoreInterval: constant.StoreInterval,
			FileStoragePath:   constant.FileStoragePath,
			StorageRestore:    constant.StorageRestore,
		},
	}
}

// Init all configs
func (c *Config) Init() (*Config, error) {
	/* old * /
	c.withFlags().WithEnv().CleanSchemes()
	err := c.LoadPrivateKey()
	/*some new*/
	c.withFlags()
	err := c.ParseEnv()

	err = c.LoadPrivateKey()
	c.CleanSchemes()

	/*new * /
		c.parseFlags()
		err := c.ParseEnv()
		if ok, er := c.maybeLoadConfig(); ok && er == nil {
			// reload flag and env after config file
			fs := flag.NewFlagSet("reload", flag.ContinueOnError)
			structflag.LoadTo(fs, "", c)
			err = fs.Parse(os.Args[1:])
			err = errors.Join(err, env.Parse(c))
		}

		err = errors.Join(err, c.LoadPrivateKey())
		c.CleanSchemes()

	/**/
	return c, err
}

// WithEnv gets ENV configs
// todo : deprecated
func (c *Config) WithEnv() *Config {
	if envVal, ok := os.LookupEnv(constant.EnvNameServerAddress); ok && envVal != "" {
		c.Address = envVal
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
	if envVal, ok := os.LookupEnv(constant.EnvNameCryptoKey); ok {
		c.CryptoKey = envVal
	}

	return c
}

// withFlags
// todo : deprecated
func (c *Config) withFlags() *Config {
	flag.StringVar(&c.Address, "a", c.Address, "Provide the address start server")
	flag.IntVar(&c.FileStoreInterval, "i", c.FileStoreInterval, "Provide the interval store (sec)")
	flag.StringVar(&c.FileStoragePath, "f", c.FileStoragePath, "Provide the file storage path")
	flag.BoolVar(&c.StorageRestore, "r", c.StorageRestore, "Restore storage at boot")
	flag.StringVar(&c.DatabaseDSN, "d", c.DatabaseDSN, "Provide the database dsn connect string")
	flag.StringVar(&c.Key, "k", c.Key, "Provide the key")
	flag.StringVar(&c.CryptoKey, "crypto-key", c.CryptoKey, "Provide the private server key for decryption")
	flag.Parse()
	return c
}

// ParseEnv gets ENV configs
func (c *Config) ParseEnv() error {
	return env.Parse(c)
}

func (c *Config) parseFlags() {
	structflag.Load(c)
	flag.Parse()
}

func (c *Config) maybeLoadConfig() (ok bool, err error) {
	if c.Config == "" && c.Config2 != "" {
		c.Config = c.Config2
	}
	if c.Config == "" {
		return
	}
	var confFile *os.File
	confFile, err = os.Open(c.Config)
	defer func() {
		err = errors.Join(err, confFile.Close())
	}()
	if err != nil {
		return
	}
	jsonParser := json.NewDecoder(confFile)
	err = jsonParser.Decode(c)
	if err != nil {
		return
	}
	ok = true
	return
}

// CleanSchemes check and repair config parameters
func (c *Config) CleanSchemes() *Config {
	for _, v := range []string{"http://", "https://"} {
		c.Address = strings.TrimPrefix(c.Address, v)
	}
	c.DatabaseDSN = strings.Trim(c.DatabaseDSN, "'")
	return c
}

func (c *WEB) GetPrivateKey() *rsa.PrivateKey {
	return c.cryptoKey
}

func (c *WEB) LoadPrivateKey() error {
	if c.CryptoKey != "" {
		b, err := os.ReadFile(c.CryptoKey)
		if err != nil {
			return err
		}

		spkiBlock, _ := pem.Decode(b)
		cert, err := x509.ParsePKCS1PrivateKey(spkiBlock.Bytes)
		if err != nil {
			return err
		}
		c.cryptoKey = cert
	}
	return nil
}
