package server

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
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
	require.NoError(suite.T(), os.RemoveAll(suite.T().TempDir()))
}

func TestConfigs(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

func (suite *ConfigTestSuite) setConfigFromMap(m map[string]any, sc ...*config.Config) (c *config.Config) {
	if len(sc) == 0 || sc[0] == nil {
		c = config.NewConfig()
	} else {
		c = sc[0]
	}
	for k, v := range m {
		switch v := v.(type) {
		case string:
			switch k {
			case "address", "-a", "ADDRESS":
				c.Address = v
			case "database_dsn", "-d", "DATABASE_DSN":
				c.DatabaseDSN = v
			case "key", "-k", "KEY":
				c.Key = v
			case "crypto_key", "-crypto-key", "CRYPTO_KEY":
				c.CryptoKey = v
			case "file_storage_path", "-f", "FILE_STORAGE_PATH":
				c.FileStoragePath = v
			case "RESTORE", "-r":
				v, err := strconv.ParseBool(v)
				require.NoError(suite.T(), err)
				c.StorageRestore = v
			case "FILE_STORE_INTERVAL":
				v, err := strconv.Atoi(v)
				require.NoError(suite.T(), err)
				c.FileStoreInterval = v
			case "config", "CONFIG":
				c.Config = v
			case "config2", "-c":
				c.Config = v
				c.Config2 = v

			}
		case bool:
			switch k {
			case "restore", "-r":
				c.StorageRestore = v
			}
		case int:
			switch k {
			case "file_store_interval", "-i":
				c.FileStoreInterval = v
			}
		}
	}
	c.CleanSchemes()
	err := c.LoadPrivateKey()
	require.NoError(suite.T(), err)
	return c
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
		config  any
		flag    map[string]any
		env     map[string]any
		name    string
		wantErr bool
	}{
		{
			name: "Default",
		},
		{
			name: "Config 1, small",
			config: map[string]any{
				"config":            cnfFile,
				"address":           "localhost:8888",
				"file_storage_path": "store.json",
			},
		},
		{
			name: "Config 2, full",
			config: map[string]any{
				"config":              cnfFile,
				"address":             "localhost:8000",
				"database_dsn":        "host=confighost port=5432 user=metric password=metric dbname=metric sslmode=disable",
				"key":                 "some-config-secret-key",
				"crypto_key":          suite.privateKey,
				"file_storage_path":   "configstore.json",
				"restore":             true,
				"file_store_interval": 100,
			},
		},
		{
			name: "Config 3, check empty's",
			config: map[string]any{
				"config":              cnfFile,
				"file_storage_path":   "",
				"restore":             false,
				"file_store_interval": 0,
			},
		},
		{
			name: "Flag 1, small",
			flag: map[string]any{
				"-i": 120,
				"-k": "some-flag-secret-key",
			},
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
		},
		{
			name: "Flag 3, check empty's",
			flag: map[string]any{
				"-f": "",
				"-r": false,
				"-i": 0,
			},
		},
		{
			name: "ENV 1, small",
			env: map[string]any{
				"FILE_STORE_INTERVAL": "250",
				"KEY":                 "some-env-secret-key",
			},
		},
		{
			name: "ENV 2, full",
			env: map[string]any{
				"ADDRESS":             "localhost:8002",
				"DATABASE_DSN":        "host=envhost port=5432 user=metric password=metric dbname=metric sslmode=disable",
				"KEY":                 "some-env-secret-key",
				"CRYPTO_KEY":          suite.privateKey,
				"FILE_STORAGE_PATH":   "envstore.json",
				"RESTORE":             "true",
				"FILE_STORE_INTERVAL": "50",
			},
		},
		{
			name: "ENV 3, check empty's",
			env: map[string]any{
				// "FILE_STORAGE_PATH":   "", // todo: env.Parse can't set empty
				"RESTORE":             "false",
				"FILE_STORE_INTERVAL": "0",
			},
		},
		{
			name: "Config and flag",
			config: map[string]any{
				"address":           "localhost:0000",
				"file_storage_path": "store.json",
			},
			flag: map[string]any{
				"-i": 120,
				"-k": "some-flag-secret-key",
				"-c": cnfFile,
			},
		},
		{
			name: "Config and flag cross",
			config: map[string]any{
				"address":           "localhost:0000",
				"file_storage_path": "store.json",
			},
			flag: map[string]any{
				"-a": "localhost:11111",
				"-k": "some-flag-secret-key",
				"-c": cnfFile,
			},
		},
		{
			name: "Config and env",
			config: map[string]any{
				"address":           "localhost:0000",
				"file_storage_path": "store.json",
			},
			env: map[string]any{
				"FILE_STORE_INTERVAL": "250",
				"KEY":                 "some-env-secret-key",
				"CONFIG":              cnfFile,
			},
		},
		{
			name: "Config and env cross",
			config: map[string]any{
				"address":           "localhost:0000",
				"file_storage_path": "store.json",
			},
			env: map[string]any{
				"ADDRESS":           "localhost:111",
				"FILE_STORAGE_PATH": "enf-store",
				"CONFIG":            cnfFile,
			},
		},
		{
			name: "flag and env",
			flag: map[string]any{
				"-a": "localhost:0000",
			},
			env: map[string]any{
				"FILE_STORE_INTERVAL": "250",
				"KEY":                 "some-env-secret-key",
			},
		},
		{
			name: "flag and env cross",
			flag: map[string]any{
				"-a": "localhost:0000",
			},
			env: map[string]any{
				"ADDRESS": "localhost:111",
			},
		},
		{
			name: "config and flag and env",
			config: map[string]any{
				"address":           "localhost:0000",
				"file_storage_path": "config-store.json",
			},
			flag: map[string]any{
				"-k": "some-flag-secret-key",
				"-c": cnfFile,
			},
			env: map[string]any{
				"FILE_STORE_INTERVAL": "250",
			},
		},
		{
			name: "config and flag and env cross",
			config: map[string]any{
				"address":           "localhost:0000",
				"file_storage_path": "config-store.json",
			},
			flag: map[string]any{
				"-k": "some-flag-secret-key",
				"-c": cnfFile,
			},
			env: map[string]any{
				"FILE_STORE_INTERVAL": "250",
				"KEY":                 "some-env-secret-key",
			},
		},
		{
			name: "bad config",
			config: `
				"address":           "localhost:0000",
				"file_storage_path": "config-store.json",
			}`,
			wantErr: true,
		},
	}

	for _, test := range tests {
		flag.CommandLine = flag.NewFlagSet(test.name, flag.ContinueOnError)
		// clean args
		os.Args = make([]string, len(test.flag)+1)
		os.Args[0] = osArgs[0]

		wantCfg := config.NewConfig()

		if test.config != nil {
			// prepare config file for test
			err := testhelpers.CreateConfigFile(cnfFile, test.config)
			require.NoError(t, err)
			if !test.wantErr {
				wantCfg = suite.setConfigFromMap(test.config.(map[string]any), wantCfg)
			}
		}
		if test.flag != nil {
			// prepare flag for test
			var i int
			for k, v := range test.flag {
				i++
				os.Args[i] = fmt.Sprintf(`%s=%v`, k, v)
			}

			wantCfg = suite.setConfigFromMap(test.flag, wantCfg)
		}
		if test.env != nil {
			// prepare env sets
			for k, v := range test.env {
				if v, ok := v.(string); ok {
					er := os.Setenv(k, v)
					require.NoError(t, er)
				}
			}
			wantCfg = suite.setConfigFromMap(test.env, wantCfg)
		}
		t.Run(test.name, func(t *testing.T) {
			var err error
			cfg := config.NewConfig()
			if test.config != nil {
				cfg.Config = cnfFile
			}

			cfg, err = cfg.Init()
			if (err != nil) != test.wantErr {
				t.Errorf("Close() error = %v, wantErr %v", err, test.wantErr)
			}
			if !test.wantErr {
				assert.Equal(t, true, reflect.DeepEqual(cfg, wantCfg), fmt.Sprintf("expected: %v\n  actual: %v", wantCfg, cfg))
			}
		})

		if test.env != nil {
			for k := range test.env {
				er := os.Unsetenv(k)
				require.NoError(t, er)
			}
		}
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
