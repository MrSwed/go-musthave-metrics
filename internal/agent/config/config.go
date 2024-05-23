package config

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"flag"
	"os"
	"strings"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/ucarion/structflag"
)

var Backoff = [3]time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

type PublicKey interface {
	ecdsa.PublicKey | rsa.PublicKey
}

type Config struct {
	Address   string `json:"address" env:"ADDRESS" flag:"a" usage:"Provide the address of the metrics collection server"`
	Key       string `json:"key" env:"KEY" flag:"k" usage:"Provide the key"`
	CryptoKey string `json:"crypto_key" env:"CRYPTO_KEY" flag:"crypto-key" usage:"Provide the public server key for encryption"`
	cryptoKey *rsa.PublicKey
	MetricLists
	ReportInterval int `json:"report_interval" env:"REPORT_INTERVAL" flag:"r" usage:"Provide the interval in seconds for send report metrics"`
	PollInterval   int `json:"poll_interval" env:"POLL_INTERVAL" flag:"p" usage:"Provide the interval in seconds for update metrics"`
	RateLimit      int `json:"rate_limit" env:"RATE_LIMIT" flag:"l" usage:"Provide the rate limit - number of concurrent outgoing requests"`
	SendSize       int `json:"send_size" env:"SEND_SIZE" flag:"s" usage:"Provide the number of metrics send at once. 0 - send all"`
}

type MetricLists struct {
	GaugesList   []string `type:"gauge" json:"gauges_list"`
	CountersList []string `type:"counter" json:"counters_list"`
}

func NewConfig() *Config {
	c := &Config{
		Address:        "localhost:8080",
		ReportInterval: 10,
		PollInterval:   2,
		Key:            "",
		RateLimit:      1,
		SendSize:       10,
	}
	c.setGaugesList()
	c.setCountersList()
	return c.CleanSchemes()
}

func (c *Config) setGaugesList(m ...string) {
	if len(m) > 0 {
		c.GaugesList = m
	} else {
		c.GaugesList = []string{
			"Alloc",
			"BuckHashSys",
			"Frees",
			"GCCPUFraction",
			"GCSys",
			"HeapAlloc",
			"HeapIdle",
			"HeapInuse",
			"HeapObjects",
			"HeapReleased",
			"HeapSys",
			"LastGC",
			"Lookups",
			"MCacheInuse",
			"MCacheSys",
			"MSpanInuse",
			"MSpanSys",
			"Mallocs",
			"NextGC",
			"NumForcedGC",
			"NumGC",
			"OtherSys",
			"PauseTotalNs",
			"StackInuse",
			"StackSys",
			"Sys",
			"TotalAlloc",
			"RandomValue",
			"TotalMemory",
			"FreeMemory",
			"CPUutilization",
		}
	}
}

func (c *Config) setCountersList(m ...string) {
	if len(m) > 0 {
		c.CountersList = m
	} else {
		c.CountersList = []string{
			"PollCount",
		}
	}
}

// Init config from flags and env
func (c *Config) Init() (err error) {
	structflag.Load(c)
	flag.Parse()
	err = env.Parse(c)
	c.CleanSchemes()
	// get key to mem
	err = errors.Join(err, c.LoadPublicKey())
	return
}

// CleanSchemes check and repair config parameters
func (c *Config) CleanSchemes() *Config {
	if !strings.HasPrefix(c.Address, "http://") && !strings.HasPrefix(c.Address, "https://") {
		c.Address = "http://" + c.Address
	}
	return c
}

func (c *Config) GetPublicKey() *rsa.PublicKey {
	return c.cryptoKey
}

func (c *Config) LoadPublicKey() error {
	if c.CryptoKey != "" {
		b, err := os.ReadFile(c.CryptoKey)
		if err != nil {
			return err
		}

		spkiBlock, _ := pem.Decode(b)
		cert, err := x509.ParseCertificate(spkiBlock.Bytes)
		if err == nil && (cert == nil || cert.PublicKey == nil) {
			err = errors.New("failed to load public key")
		}
		if err != nil {
			return err
		}
		c.cryptoKey = cert.PublicKey.(*rsa.PublicKey)
	}
	return nil
}
