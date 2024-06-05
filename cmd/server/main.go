package main

import (
	"context"
	_ "net/http/pprof"
	"os/signal"
	"syscall"

	"go-musthave-metrics/internal/server/app"
	"go-musthave-metrics/internal/server/config"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

var buildVersion string
var buildDate string
var buildCommit string

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()
	cfg, err := config.NewConfig().Init()
	if err != nil {
		panic(err)
	}

	log, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	appHandler := app.NewApp(ctx, stop, app.BuildMetadata{
		Version: buildVersion,
		Date:    buildDate,
		Commit:  buildCommit,
	}, cfg, log)

	appHandler.Run()
	appHandler.Stop()
}
