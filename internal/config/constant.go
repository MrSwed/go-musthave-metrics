package config

import (
	"embed"
)

const (
	ServerAddress = "localhost:8080"

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
