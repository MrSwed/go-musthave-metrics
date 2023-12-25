package constants

import "embed"

const (
	UpdateRoute = "/update"
	ValueRoute  = "/value"

	MetricTypeGauge   = "gauge"
	MetricTypeCounter = "counter"
)

var (
	//go:embed template/list_tpl.html
	MetricListTpl embed.FS
)
