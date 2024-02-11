package config

import (
	"embed"
)

const (
	ServerShutdownTimeout = 30

	serverAddress   = "localhost:8080"
	storeInterval   = 300
	fileStoragePath = "/tmp/metrics-db.json"
	storageRestore  = true

	envNameServerAddress   = "ADDRESS"
	envNameFileStoragePath = "FILE_STORAGE_PATH"
	envNameStoreInterval   = "STORE_INTERVAL"
	envNameRestore         = "RESTORE"
	envNameDBDSN           = "DATABASE_DSN"

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

	RetriesOnErr = [3]int{1, 3, 5}
)
