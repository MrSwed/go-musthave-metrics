package config

import (
	"embed"
)

const (
	serverAddress   = "localhost:8080"
	storeInterval   = 300
	fileStoragePath = "/tmp/metrics-db.json"
	storageRestore  = true

	envNameServerAddress   = "SERVER_ADDRESS"
	envNameFileStoragePath = "FILE_STORAGE_PATH"
	envNameStoreInterval   = "STORE_INTERVAL"
	envNameRestore         = "RESTORE"

	UpdateRoute      = "/update"
	ValueRoute       = "/value"
	MetricTypeParam  = "metricType"
	MetricNameParam  = "metricName"
	MetricValueParam = "metricValue"

	MetricTypeGauge   = "gauge"
	MetricTypeCounter = "counter"
)

var (
	//go:embed template/list_tpl.html
	MetricListTpl embed.FS
)
