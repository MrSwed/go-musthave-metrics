package main

import (
	"context"
	"go-musthave-metrics/internal/agent/app"
)

var buildVersion string
var buildDate string
var buildCommit string

func main() {
	app.RunApp(context.Background(), nil,
		app.BuildMetadata{Version: buildVersion, Date: buildDate, Commit: buildCommit})
}
