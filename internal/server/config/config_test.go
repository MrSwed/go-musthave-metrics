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
	"address": "localhost:8088",
	"restore": true,
	"file_store_interval": 1,
	"file_storage_path": "/tmp/metrics",
	"database_dsn": "host=localhost port=5432 user=metric password=metric dbname=go_musthave_metrics sslmode=disable",
	"crypto_key": "./cmd/.cert/server.key",
	"key": "config-some-secret-key"
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
		assert.Equal(t, "localhost:8088", c.Address)
		assert.Equal(t, true, c.StorageRestore)
		assert.Equal(t, 1, c.FileStoreInterval)
		assert.Equal(t, "/tmp/metrics", c.FileStoragePath)
		assert.Equal(t, "host=localhost port=5432 user=metric password=metric dbname=go_musthave_metrics sslmode=disable", c.DatabaseDSN)
		assert.Equal(t, "./cmd/.cert/server.key", c.CryptoKey)
	})
}
