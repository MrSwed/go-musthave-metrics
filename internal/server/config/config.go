package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"flag"
	"net"
	"os"
	"strings"

	"go-musthave-metrics/internal/server/constant"
	"go-musthave-metrics/pkg/structflag"

	"github.com/caarlos0/env/v11"
)

// StorageConfig file storage configs
type StorageConfig struct {
	FileStoragePath   string `env:"FILE_STORAGE_PATH" json:"file_storage_path" flag:"f" usage:"Provide the file storage path"`
	StorageRestore    bool   `env:"RESTORE" json:"restore" flag:"r" usage:"Provide the file storage path"`
	FileStoreInterval int    `env:"FILE_STORE_INTERVAL" json:"file_store_interval" flag:"i" usage:"Provide the interval in seconds"`
}

// WEB  config
type WEB struct {
	cryptoKey     *rsa.PrivateKey
	Key           string `env:"KEY" json:"key" flag:"k" usage:"Private theKey"`
	CryptoKey     string `env:"CRYPTO_KEY" json:"crypto_key" flag:"crypto-key" usage:"Provide the private server key for decryption"`
	TrustedSubnet string `env:"TRUSTED_SUBNET" json:"trusted_subnet" flag:"t" usage:"Provide the trusted subnet"`
}

type GRPC struct {
	GRPCAddress string `env:"GRPC_ADDRESS" json:"grpc_address"  flag:"g" usage:"Provide the grpc service address"`
	GRPCToken   string `env:"GRPC_TOKEN" json:"grpc_token"  flag:"grpc_token" usage:"Provide the grpc service token"`
}

// Config all configs
type Config struct {
	Address     string `env:"ADDRESS" json:"address"  flag:"a" usage:"Provide the address start server"`
	DatabaseDSN string `env:"DATABASE_DSN" json:"database_dsn" flag:"d" usage:"Provide the database dsn connect string"`
	Config      string `json:"-" env:"CONFIG" flag:"config" usage:"Provide file with config"`
	Config2     string `json:"-" env:"-" flag:"c" usage:"same as -config"` // ?
	WEB
	GRPC
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
		GRPC: GRPC{
			GRPCAddress: constant.GRPCAddress,
		},
	}
}

// Init all configs
func (c *Config) Init() (*Config, error) {
	c.parseFlags()
	err := c.ParseEnv()
	if ok, er := c.maybeLoadConfig(); ok && er == nil {
		// reload flag and env after config file
		fs := flag.NewFlagSet("reload", flag.ContinueOnError)
		structflag.LoadTo(fs, "", c)
		err = errors.Join(err,
			fs.Parse(os.Args[1:]),
			env.Parse(c),
		)
	} else {
		err = errors.Join(err, er)
	}
	if c.TrustedSubnet != "" {
		if _, _, er := net.ParseCIDR(c.TrustedSubnet); er != nil {
			err = errors.Join(err, er)
		}
	}

	err = errors.Join(err, c.LoadPrivateKey())
	c.CleanSchemes()

	return c, err
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
	if err != nil {
		return
	}
	defer func() {
		err = errors.Join(err, confFile.Close())
	}()
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
