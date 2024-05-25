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
	ctx                     context.Context
	existKey, existWrongKey string
}

func (suite *ConfigTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	suite.existWrongKey = suite.T().TempDir() + "/testPrivate.key"
	suite.existKey = suite.T().TempDir() + "/testPublic.crt"
	testhelpers.CreateCertificates(suite.existKey, suite.existWrongKey)
}

func (suite *ConfigTestSuite) TearDownSuite() {
	_ = os.Remove(suite.existKey)
	_ = os.Remove(suite.existWrongKey)
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
		env    map[string]any
		want   *config.Config
		name   string
	}{
		{
			name: "Default",
			want: config.NewConfig(),
		},
		{
			name: "Config",
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
			name: "Flag",
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
		}
		t.Run(test.name, func(t *testing.T) {
			var err error
			cfg := config.NewConfig()
			if test.config != nil {
				cfg.Config = cnfFile
			}
			cfg, err = cfg.Init()
			assert.NoError(t, err)
			// cm := reflect.DeepEqual(cfg, test.want)
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
		file: suite.existKey,
		ok:   true,
	},
		{
			name: "Not exist key",
			file: "/someNoteExist.key",
			ok:   false,
		},
		{
			name: "Wrong key",
			file: suite.existWrongKey,
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
