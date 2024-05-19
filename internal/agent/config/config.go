package config

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"flag"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/MrSwed/go-musthave-metrics/internal/agent/constant"
)

var Backoff = [3]time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

type PublicKey interface {
	ecdsa.PublicKey | rsa.PublicKey
}

type Config struct {
	ServerAddress string
	Key           string
	CryptoKey     string
	cryptoKey     *rsa.PublicKey
	MetricLists
	ReportInterval int
	PollInterval   int
	RateLimit      int
	SendSize       int
}

type MetricLists struct {
	GaugesList   []string `type:"gauge"`
	CountersList []string `type:"counter"`
}

func NewConfig() *Config {
	c := &Config{
		ServerAddress:  "localhost:8080",
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

func (c *Config) parseFlags() {
	flag.StringVar(&c.ServerAddress, "a", c.ServerAddress, "Provide the address of the metrics collection server")
	flag.IntVar(&c.ReportInterval, "r", c.ReportInterval, "Provide the interval in seconds for send report metrics")
	flag.IntVar(&c.PollInterval, "p", c.PollInterval, "Provide the interval in seconds for update metrics")
	flag.StringVar(&c.Key, "k", c.Key, "Provide the key")
	flag.IntVar(&c.RateLimit, "l", c.RateLimit, "Provide the rate limit - number of concurrent outgoing requests")
	flag.IntVar(&c.SendSize, "s", c.SendSize, "Provide the number of metrics send at once. 0 - send all")
	flag.StringVar(&c.CryptoKey, "crypto-key", c.CryptoKey, "Provide the public server key for encryption")
	flag.Parse()
}

func (c *Config) getEnv() {
	addressEnv, reportIntervalEnv, pollIntervalEnv, key, rateLimit, cryptoKey :=
		os.Getenv(constant.EnvNameServerAddress),
		os.Getenv(constant.EnvNameReportInterval),
		os.Getenv(constant.EnvNamePollInterval),
		os.Getenv(constant.EnvNameKey),
		os.Getenv(constant.EnvNameRateLimit),
		os.Getenv(constant.EnvNameCryptoKey)
	if addressEnv != "" {
		c.ServerAddress = addressEnv
	}
	if reportIntervalEnv != "" {
		if v, err := strconv.Atoi(reportIntervalEnv); err == nil {
			c.ReportInterval = v
		}
	}
	if pollIntervalEnv != "" {
		if v, err := strconv.Atoi(pollIntervalEnv); err == nil {
			c.PollInterval = v
		}
	}
	if key != "" {
		c.Key = key
	}
	if cryptoKey != "" {
		c.CryptoKey = cryptoKey
	}
	if rateLimit != "" {
		if v, err := strconv.Atoi(rateLimit); err == nil {
			c.RateLimit = v
		}
	}
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
	c.parseFlags()
	c.getEnv()
	c.CleanSchemes()
	// get key to mem
	err = c.LoadPublicKey()
	return
}

// CleanSchemes check and repair config parameters
func (c *Config) CleanSchemes() *Config {
	if !strings.HasPrefix(c.ServerAddress, "http://") && !strings.HasPrefix(c.ServerAddress, "https://") {
		c.ServerAddress = "http://" + c.ServerAddress
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
