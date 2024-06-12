package agent

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"

	"go-musthave-metrics/internal/agent/config"
	helper "go-musthave-metrics/tests"

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
	suite.privateKey = filepath.Join(suite.T().TempDir(), "testPrivate.key")
	suite.publicKey = filepath.Join(suite.T().TempDir(), "testPublic.crt")
	helper.CreateCertificates(suite.privateKey, suite.publicKey)
}

func (suite *ConfigTestSuite) TearDownSuite() {}

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
		case int:
			switch k {
			case "report_interval", "-r":
				c.ReportInterval = v
			case "poll_interval", "-p":
				c.PollInterval = v
			case "rate_limit", "-l":
				c.RateLimit = v
			case "send_size", "-s":
				c.SendSize = v
			}
		case string:
			switch k {
			case "address", "-a", "ADDRESS":
				c.Address = v
			case "key", "-k", "KEY":
				c.Key = v
			case "crypto_key", "-crypto-key", "CRYPTO_KEY":
				c.CryptoKey = v
			case "config", "CONFIG":
				c.Config = v
			case "config2", "-c":
				c.Config = v
				c.Config2 = v
			case "REPORT_INTERVAL":
				v, err := strconv.Atoi(v)
				require.NoError(suite.T(), err)
				c.ReportInterval = v
			case "POLL_INTERVAL":
				v, err := strconv.Atoi(v)
				require.NoError(suite.T(), err)
				c.PollInterval = v
			case "RATE_LIMIT":
				v, err := strconv.Atoi(v)
				require.NoError(suite.T(), err)
				c.RateLimit = v
			case "SEND_SIZE":
				v, err := strconv.Atoi(v)
				require.NoError(suite.T(), err)
				c.SendSize = v
			}
		}
	}
	c.SetDefaultMetrics()
	c.CleanSchemes()
	err := c.LoadPublicKey()
	require.NoError(suite.T(), err)
	return c
}

func (suite *ConfigTestSuite) TestInit() {
	t := suite.T()

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	cnfFile := filepath.Join(t.TempDir(), "config.json")

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
				"address":         "localhost:8010",
				"report_interval": 100,
				"config":          cnfFile,
			},
		},
		{
			name: "Config 2, full",
			config: map[string]any{
				"config":          cnfFile,
				"address":         "localhost:8011",
				"report_interval": 101,
				"poll_interval":   11,
				"rate_limit":      11,
				"send_size":       11,
				"key":             "some-config-secret-key",
				"crypto_key":      suite.publicKey,
			},
		},
		{
			name: "Config 3, check empty's",
			config: map[string]any{
				"config":          cnfFile,
				"address":         "",
				"report_interval": 0,
			},
		},
		{
			name: "Flag 1, small",
			flag: map[string]any{
				"-a": "localhost:8021",
				"-r": 201,
			},
		},
		{
			name: "Flag 2, full",
			flag: map[string]any{
				"-a":          "localhost:8022",
				"-r":          203,
				"-p":          22,
				"-l":          22,
				"-s":          22,
				"-k":          "some-flag-secret-key",
				"-crypto-key": suite.publicKey,
			},
		},
		{
			name: "Flag 3, check empty's",
			flag: map[string]any{
				"-a": "",
				"-r": 0,
			},
		},
		{
			name: "ENV 1, small",
			env: map[string]any{
				"ADDRESS":         "localhost:8033",
				"REPORT_INTERVAL": "301",
			},
		},
		{
			name: "ENV 2, full",
			env: map[string]any{
				"ADDRESS":         "localhost:8003",
				"REPORT_INTERVAL": "301",
				"POLL_INTERVAL":   "31",
				"RATE_LIMIT":      "31",
				"SEND_SIZE":       "31",
				"KEY":             "some-env-secret-key",
				"CRYPTO_KEY":      suite.publicKey,
			},
		},
		{
			name: "ENV 3, check empty's",
			env: map[string]any{
				// "ADDRESS":         "", // todo: env.Parse can't set empty
				"REPORT_INTERVAL": "0",
			},
		},
		{
			name: "Config and flag",
			config: map[string]any{
				"address": "localhost:8110",
			},
			flag: map[string]any{
				"-r": 120,
				"-c": cnfFile,
			},
		},
		{
			name: "Config and flag cross",
			config: map[string]any{
				"address": "localhost:0000",
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
				"address": "localhost:0000",
			},
			env: map[string]any{
				"REPORT_INTERVAL": "250",
				"KEY":             "some-env-secret-key",
				"CONFIG":          cnfFile,
			},
		},
		{
			name: "Config and env cross",
			config: map[string]any{
				"address": "localhost:0000",
			},
			env: map[string]any{
				"ADDRESS": "localhost:111",
				"CONFIG":  cnfFile,
			},
		},
		{
			name: "flag and env",
			flag: map[string]any{
				"-a": "localhost:0000",
			},
			env: map[string]any{
				"REPORT_INTERVAL": "250",
				"KEY":             "some-env-secret-key",
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
				"address": "localhost:0000",
			},
			flag: map[string]any{
				"-k": "some-flag-secret-key",
				"-c": cnfFile,
			},
			env: map[string]any{
				"REPORT_INTERVAL": "250",
			},
		},
		{
			name: "config and flag and env cross",
			config: map[string]any{
				"address":         "localhost:0000",
				"report_interval": 11,
			},
			flag: map[string]any{
				"-k": "some-flag-secret-key",
				"-c": cnfFile,
			},
			env: map[string]any{
				"REPORT_INTERVAL": "50",
				"KEY":             "some-env-secret-key",
			},
		},
		{
			name: "bad config file",
			config: ` 
				"address":         "localhost:0000",
				"report_interval": 11,
			}`,
			wantErr: true,
		},
	}

	for _, test := range tests {
		flag.CommandLine = flag.NewFlagSet(test.name, flag.ContinueOnError)
		os.Args = make([]string, len(test.flag)+1)
		os.Args[0] = oldArgs[0]

		wantCfg := config.NewConfig()

		if test.config != nil {
			// prepare config file for test
			err := helper.CreateConfigFile(cnfFile, test.config)
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
			err = cfg.Init()
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

func (suite *ConfigTestSuite) TestLoadPublicKey() {
	t := suite.T()

	tests := []struct {
		file string
		name string
		ok   bool
	}{{
		name: "Exist key",
		file: suite.publicKey,
		ok:   true,
	},
		{
			name: "Not exist key",
			file: "/someNoteExist.key",
			ok:   false,
		},
		{
			name: "Wrong key",
			file: suite.privateKey,
			ok:   false,
		},
	}
	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {
			cfg := config.NewConfig()

			cfg.CryptoKey = test.file
			err := cfg.LoadPublicKey()
			if (err == nil) != test.ok {
				t.Errorf("Incorrect error result (did fail? %v, expected: %v) err %v", err == nil, test.ok, err)
			}
		})
	}

}
