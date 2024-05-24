package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_maybeLoadConfig(t *testing.T) {
	const configStr = `
{
	"address": "http://localhost:3005",
	"report_interval": 5,
	"poll_interval": 1,
	"rate_limit": 5,
	"send_size": 5,
	"crypto_key": "./cmd/.cert/server.crt"
}
`
	configFile := t.TempDir() + "config.json"
	file, err := os.Create(configFile)
	require.NoError(t, err)
	defer func() {
		err = file.Close()
		require.NoError(t, err)
		err = os.Remove(file.Name())
		require.NoError(t, err)
	}()
	_, err = file.WriteString(configStr)
	require.NoError(t, err)

	t.Run("load", func(t *testing.T) {
		c := &Config{
			Config: configFile,
		}
		gotOk, err := c.maybeLoadConfig()
		assert.NoError(t, err)
		assert.True(t, gotOk)
		assert.Equal(t, "http://localhost:3005", c.Address)
		assert.Equal(t, 5, c.ReportInterval)
		assert.Equal(t, 1, c.PollInterval)
		assert.Equal(t, 5, c.RateLimit)
		assert.Equal(t, 5, c.SendSize)
		assert.Equal(t, "./cmd/.cert/server.crt", c.CryptoKey)
	})
}
