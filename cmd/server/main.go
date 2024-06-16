package main

import (
	"context"
	"go-musthave-metrics/internal/server/app"
	_ "net/http/pprof"

	_ "github.com/lib/pq"
)

var buildVersion string
var buildDate string
var buildCommit string

func main() {
	app.RunApp(context.Background(), nil, nil,
		app.BuildMetadata{Version: buildVersion, Date: buildDate, Commit: buildCommit})
}
