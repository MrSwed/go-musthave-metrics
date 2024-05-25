package server

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/MrSwed/go-musthave-metrics/internal/server/config"
	"github.com/MrSwed/go-musthave-metrics/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ConfigTestSuite struct {
	suite.Suite
	ctx                   context.Context
	publicKey, privateKey string
}

func (suite *ConfigTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	suite.privateKey = suite.T().TempDir() + "/testPrivate.key"
	suite.publicKey = suite.T().TempDir() + "/testPublic.crt"
	testhelpers.CreateCertificates(suite.privateKey, suite.publicKey)
}

func (suite *ConfigTestSuite) TearDownSuite() {
	_ = os.Remove(suite.publicKey)
	_ = os.Remove(suite.privateKey)
}

func TestConfigs(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

func (suite *ConfigTestSuite) TestInit() {
	t := suite.T()

	osArgs := make([]string, len(os.Args))
	copy(osArgs, os.Args)
	// do not use t.Parallel with one config file
	cnfFile := filepath.Join(t.TempDir(), "config.json")
	defer func() {
		_ = os.Remove(cnfFile)
		copy(os.Args, osArgs)
	}()

	tests := []struct {
		config map[string]any
		flag   map[string]any
		env    map[string]string
		want   *config.Config
		name   string
	}{
		{
			name: "Default",
			want: config.NewConfig(),
		},
		{
			name: "Config 1, small",
			config: map[string]any{
				"address":           "localhost:8888",
				"file_storage_path": "store.json",
			},
			want: func() (c *config.Config) {
				c = config.NewConfig()
				c.Address = "localhost:8888"
				c.StorageConfig.FileStoragePath = "store.json"
				c.Config = cnfFile
				return c
			}(),
		},
		{
			name: "Config 2, full",
			config: map[string]any{
				"address":             "localhost:8000",
				"database_dsn":        "host=confighost port=5432 user=metric password=metric dbname=metric sslmode=disable",
				"key":                 "some-config-secret-key",
				"crypto_key":          suite.privateKey,
				"file_storage_path":   "configstore.json",
				"restore":             true,
				"file_store_interval": 100,
			},
			want: func() (c *config.Config) {
				c = config.NewConfig()
				c.Address = "localhost:8000"
				c.DatabaseDSN = "host=confighost port=5432 user=metric password=metric dbname=metric sslmode=disable"
				c.Key = "some-config-secret-key"
				c.CryptoKey = suite.privateKey
				c.FileStoragePath = "configstore.json"
				c.FileStoreInterval = 100
				c.StorageRestore = true
				c.Config = cnfFile
				err := c.LoadPrivateKey()
				require.NoError(t, err)
				return c
			}(),
		},
		{
			name: "Flag 1, small",
			flag: map[string]any{
				"-i": 120,
				"-k": "some-flag-secret-key",
			},
			want: func() (c *config.Config) {
				c = config.NewConfig()
				c.StorageConfig.FileStoreInterval = 120
				c.WEB.Key = "some-flag-secret-key"
				return c
			}(),
		},
		{
			name: "Flag 2, full",
			flag: map[string]any{
				"-a":          "localhost:8001",
				"-d":          "host=flaghost port=5432 user=metric password=metric dbname=metric sslmode=disable",
				"-k":          "some-flag-secret-key",
				"-crypto-key": suite.privateKey,
				"-f":          "flagstore.json",
				"-r":          true,
				"-i":          200,
			},
			want: func() (c *config.Config) {
				c = config.NewConfig()
				c.Address = "localhost:8001"
				c.DatabaseDSN = "host=flaghost port=5432 user=metric password=metric dbname=metric sslmode=disable"
				c.Key = "some-flag-secret-key"
				c.CryptoKey = suite.privateKey
				c.FileStoragePath = "flagstore.json"
				c.StorageRestore = true
				c.FileStoreInterval = 200
				err := c.LoadPrivateKey()
				require.NoError(t, err)
				return c
			}(),
		},
		{
			name: "ENV 1, small",
			env: map[string]string{
				"FILE_STORE_INTERVAL": "250",
				"KEY":                 "some-env-secret-key",
			},
			want: func() (c *config.Config) {
				c = config.NewConfig()
				c.StorageConfig.FileStoreInterval = 250
				c.WEB.Key = "some-env-secret-key"
				return c
			}(),
		},
		{
			name: "ENV 2, full",
			env: map[string]string{
				"ADDRESS":             "localhost:8002",
				"DATABASE_DSN":        "host=envhost port=5432 user=metric password=metric dbname=metric sslmode=disable",
				"KEY":                 "some-env-secret-key",
				"CRYPTO_KEY":          suite.privateKey,
				"FILE_STORAGE_PATH":   "envstore.json",
				"RESTORE":             "true",
				"FILE_STORE_INTERVAL": "50",
			},
			want: func() (c *config.Config) {
				c = config.NewConfig()
				c.Address = "localhost:8002"
				c.DatabaseDSN = "host=envhost port=5432 user=metric password=metric dbname=metric sslmode=disable"
				c.Key = "some-env-secret-key"
				c.CryptoKey = suite.privateKey
				c.FileStoragePath = "envstore.json"
				c.StorageRestore = true
				c.FileStoreInterval = 50
				err := c.LoadPrivateKey()
				require.NoError(t, err)
				return c
			}(),
		},
	}

	for _, test := range tests {
		flag.CommandLine = flag.NewFlagSet(test.name, flag.ContinueOnError)
		// clean args
		os.Args = make([]string, len(test.flag)+1)
		os.Args[0] = osArgs[0]

		if test.config != nil {
			// prepare config file for test
			err := testhelpers.CreateConfigFile(cnfFile, test.config)
			require.NoError(t, err)
		}
		if test.flag != nil {
			// prepare flag for test
			var i int
			for k, v := range test.flag {
				i++
				os.Args[i] = fmt.Sprintf(`%s=%v`, k, v)
			}
		}
		if test.env != nil {
			// prepare env sets
			for k, v := range test.env {
				er := os.Setenv(k, v)
				require.NoError(t, er)
			}
		}
		t.Run(test.name, func(t *testing.T) {
			var err error
			cfg := config.NewConfig()
			if test.config != nil {
				cfg.Config = cnfFile
			}
			cfg, err = cfg.Init()
			assert.NoError(t, err)
			assert.Equal(t, true, reflect.DeepEqual(cfg, test.want), fmt.Sprintf("expected: %v\n  actual: %v", test.want, cfg))
		})
	}
}

func (suite *ConfigTestSuite) TestLoadPrivateKey() {
	t := suite.T()

	tests := []struct {
		file string
		name string
		ok   bool
	}{{
		name: "Exist key",
		file: suite.privateKey,
		ok:   true,
	},
		{
			name: "Not exist key",
			file: "/someNoteExist.key",
			ok:   false,
		},
		{
			name: "Wrong key",
			file: suite.publicKey,
			ok:   false,
		},
	}
	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {
			cfg := config.NewConfig()

			cfg.CryptoKey = test.file
			err := cfg.LoadPrivateKey()
			if (err == nil) != test.ok {
				t.Errorf("Incorrect error result (did fail? %v, expected: %v) err %v", err == nil, test.ok, err)
			}
		})
	}
}
