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

	GRPCAddress = ":3200"

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

	HeaderSignKey = "HashSHA256"
	HeaderXRealIP = "X-Real-IP"
)

var (
	//go:embed template/list_tpl.html
	MetricListTpl embed.FS

	Backoff = [3]time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}
)
