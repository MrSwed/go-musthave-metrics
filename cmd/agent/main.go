package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/MrSwed/go-musthave-metrics/internal/agent/app"
	"github.com/MrSwed/go-musthave-metrics/internal/agent/config"
)

var buildVersion string
var buildDate string
var buildCommit string

func main() {
	cfg := config.NewConfig()

	if err := cfg.Init(); err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	appHandler := app.NewApp(ctx, cfg, app.BuildMetadata{
		Version: buildVersion,
		Date:    buildDate,
		Commit:  buildCommit,
	})

	appHandler.Run()
	appHandler.Stop()
}
