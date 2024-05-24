package agent

import (
	"context"
	"os"
	"testing"

	"github.com/MrSwed/go-musthave-metrics/internal/agent/config"
	testHelpers "github.com/MrSwed/go-musthave-metrics/tests"
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
	testHelpers.CreateCertificates(suite.existWrongKey, suite.existKey)
}

func (suite *ConfigTestSuite) TearDownSuite() {
	_ = os.Remove(suite.existKey)
	_ = os.Remove(suite.existWrongKey)
}

func TestConfigs(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

func (suite *ConfigTestSuite) TestLoadPublicKey() {
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
			err := cfg.LoadPublicKey()
			if (err == nil) != test.ok {
				t.Errorf("Incorrect error result (did fail? %v, expected: %v) err %v", err == nil, test.ok, err)
			}
		})
	}

}
