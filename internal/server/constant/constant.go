package constant

import (
	"embed"
	"time"
)

const (
	ServerShutdownTimeout  = 30
	ServerOperationTimeout = 30

	ServerAddress   = "localhost:8080"
	StoreInterval   = 300
	FileStoragePath = "/tmp/metrics-db.json"
	StorageRestore  = true

	EnvNameServerAddress   = "ADDRESS"
	EnvNameFileStoragePath = "FILE_STORAGE_PATH"
	EnvNameStoreInterval   = "STORE_INTERVAL"
	EnvNameRestore         = "RESTORE"
	EnvNameDBDSN           = "DATABASE_DSN"
	EnvNameKey             = "KEY"

	UpdateRoute      = "/update"
	UpdatesRoute     = "/updates"
	ValueRoute       = "/value"
	MetricTypeParam  = "metricType"
	MetricNameParam  = "metricName"
	MetricValueParam = "metricValue"

	MetricTypeGauge   = "gauge"
	MetricTypeCounter = "counter"

	DBTableNameGauges   = "gauges"
	DBTableNameCounters = "counters"
)

var (
	//go:embed template/list_tpl.html
	MetricListTpl embed.FS

	Backoff = [3]time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}
)
