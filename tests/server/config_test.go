package server

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
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

	// do not use t.Parallel with one config file
	cnfFile := filepath.Join(t.TempDir(), "config.json")
	defer func() {
		_ = os.Remove(cnfFile)
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
	}

	for _, test := range tests {
		if test.config != nil {
			// prepare config file for test
			err := testhelpers.CreateConfigFile(cnfFile, test.config)
			require.NoError(t, err)
		}
		if test.flag != nil {
			// prepare flag for test
		}
		if test.env != nil {
			// prepare env sets
		}
		t.Run(test.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(t.Name(), flag.ContinueOnError)
			var err error
			cfg := config.NewConfig()
			if test.config != nil {
				cfg.Config = cnfFile
			}
			cfg, err = cfg.Init()
			if err != nil {
				if strings.HasPrefix(err.Error(), "flag provided but not defined") {
					err = nil
				}
			}
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
