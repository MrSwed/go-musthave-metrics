package agent

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	servApp "go-musthave-metrics/internal/server/app"
	servConfig "go-musthave-metrics/internal/server/config"
	"log"
	"math/rand"
	"net"
	"os"
	"testing"
	"time"

	"go-musthave-metrics/internal/agent/app"
	"go-musthave-metrics/internal/agent/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func Test_app_Run(t *testing.T) {
	oldArgs := os.Args[:]

	type fields struct {
		cfg       *config.Config
		buildInfo app.BuildMetadata
		sCfg      *servConfig.Config
	}
	tests := []struct {
		name             string
		fields           fields
		wantStrings      []string
		doNotWantStrings []string
	}{{
		name: "Agent without server",
		fields: func() fields {
			cfg := config.NewConfig()
			cfg.ReportInterval = 2
			cfg.PollInterval = 1
			return fields{
				cfg: cfg,
				buildInfo: app.BuildMetadata{
					Version: "1.0-testing",
					Date:    "24.05.24",
					Commit:  "444333",
				},
			}
		}(),
		wantStrings: []string{
			`BuildVersion: 1.0-testing`,
			`BuildDate: 24.05.24`,
			`BuildCommit: 444333`,
			`Url for collect metric: http://localhost:8080/updates`,
			`Report interval: 2`,
			`Poll interval: 1`,
			`Rate limit: 1`,
			`Number of metrics at once: 10`,
			`Metric names count: 32`,
			`daemon started: collect runtime metrics with interval 1`,
			`daemon started: collect psutil metrics with interval 1`,
			`daemon started: send metrics interval 2`,
			`Collect runtime metrics`,
			`Collect psutil metrics`,
			`PSUtil metrics collector is stopped`,
			`Runtime metrics collector is stopped`,
			`ctx done, do not try more`,
			`second before next try`,
			`Agent stopped`,
		},
	},
		{
			name: "Agent grpc without server",
			fields: func() fields {
				cfg := config.NewConfig()
				cfg.ReportInterval = 2
				cfg.PollInterval = 1
				cfg.GRPCAddress = net.JoinHostPort("localhost", fmt.Sprintf("%d", rand.Intn(200)+20000))
				return fields{
					cfg: cfg,
					buildInfo: app.BuildMetadata{
						Version: "1.0-testing",
						Date:    "24.05.24",
						Commit:  "444333",
					},
				}
			}(),
			wantStrings: []string{
				`BuildVersion: 1.0-testing`,
				`BuildDate: 24.05.24`,
				`BuildCommit: 444333`,
				`Url for collect metric: http://localhost:8080/updates`,
				`Report interval: 2`,
				`Poll interval: 1`,
				`Rate limit: 1`,
				`Number of metrics at once: 10`,
				`Metric names count: 32`,
				`daemon started: collect runtime metrics with interval 1`,
				`daemon started: collect psutil metrics with interval 1`,
				`daemon started: send metrics interval 2`,
				`Collect runtime metrics`,
				`Collect psutil metrics`,
				`PSUtil metrics collector is stopped`,
				`Runtime metrics collector is stopped`,
				`transport: Error while dialing`,
				`Metrics sender is stopped`,
				`Agent stopped`,
			},
		},
		{
			name: "Agent and server",
			fields: func() fields {

				servCfg := servConfig.NewConfig()
				servCfg.StorageConfig.FileStoragePath = ""
				servCfg.Address = net.JoinHostPort("localhost", fmt.Sprintf("%d", rand.Intn(200)+20000))
				servCfg.GRPCAddress = ""
				// servCfg.GRPCAddress = net.JoinHostPort("", fmt.Sprintf("%d", rand.Intn(200)+30000))

				cfg := config.NewConfig()
				cfg.ReportInterval = 2
				cfg.PollInterval = 1
				cfg.Address = servCfg.Address

				return fields{
					cfg:  cfg,
					sCfg: servCfg,
					buildInfo: app.BuildMetadata{
						Version: "1.1-testing",
						Date:    "24.05.24",
						Commit:  "4444444",
					},
				}
			}(),
			wantStrings: []string{
				`BuildVersion: 1.1-testing`,
				`BuildDate: 24.05.24`,
				`BuildCommit: 4444444`,
				`Url for collect metric: http://localhost`,
				`Report interval: 2`,
				`Poll interval: 1`,
				`Rate limit: 1`,
				`Number of metrics at once: 10`,
				`Metric names count: 32`,
				`daemon started: collect runtime metrics with interval 1`,
				`daemon started: collect psutil metrics with interval 1`,
				`daemon started: send metrics interval 2`,
				`Collect runtime metrics`,
				`Collect psutil metrics`,
				`PSUtil metrics collector is stopped`,
				`Runtime metrics collector is stopped`,
				`Metrics sender is stopped`,
				`metrics sent`,
				`Agent stopped`,
			},
		},
		{
			name: "Agent and server GRPC",
			fields: func() fields {

				servCfg := servConfig.NewConfig()
				servCfg.StorageConfig.FileStoragePath = ""
				servCfg.Address = net.JoinHostPort("localhost", fmt.Sprintf("%d", rand.Intn(200)+20000))
				servCfg.GRPCAddress = net.JoinHostPort("", fmt.Sprintf("%d", rand.Intn(200)+30000))
				servCfg.GRPCToken = "secretToken"

				cfg := config.NewConfig()
				cfg.ReportInterval = 2
				cfg.PollInterval = 1
				cfg.GRPCAddress = servCfg.GRPCAddress
				cfg.GRPCToken = servCfg.GRPCToken

				return fields{
					cfg:  cfg,
					sCfg: servCfg,
					buildInfo: app.BuildMetadata{
						Version: "1.1-testing",
						Date:    "24.05.24",
						Commit:  "4444444",
					},
				}
			}(),
			wantStrings: []string{
				`BuildVersion: 1.1-testing`,
				`BuildDate: 24.05.24`,
				`BuildCommit: 4444444`,
				`Url for collect metric: http://localhost`,
				`Report interval: 2`,
				`Poll interval: 1`,
				`Rate limit: 1`,
				`Number of metrics at once: 10`,
				`Metric names count: 32`,
				`daemon started: collect runtime metrics with interval 1`,
				`daemon started: collect psutil metrics with interval 1`,
				`daemon started: send metrics interval 2`,
				`Collect runtime metrics`,
				`Collect psutil metrics`,
				`PSUtil metrics collector is stopped`,
				`Runtime metrics collector is stopped`,
				`Metrics sender is stopped`,
				`metrics sent`,
				`Agent stopped`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			flag.CommandLine = flag.NewFlagSet(tt.name, flag.ContinueOnError)
			os.Args = oldArgs[:1]
			defer func() { os.Args = oldArgs }()

			require.NoError(t, tt.fields.cfg.Init())

			var buf bytes.Buffer
			log.SetOutput(&buf)
			defer func() {
				log.SetOutput(os.Stderr)
			}()
			var sApp *servApp.App
			if tt.fields.sCfg != nil {
				go servApp.RunApp(ctx, tt.fields.sCfg,
					zap.NewNop(), servApp.BuildMetadata{Version: "testing..", Date: time.Now().String(), Commit: ""})
			}

			appHandler := app.NewApp(ctx, tt.fields.cfg, tt.fields.buildInfo)
			appHandler.Run()
			appHandler.Stop()

			if sApp != nil {
				sApp.Stop()
			}
			t.Log(buf.String())
			for i := 0; i < len(tt.wantStrings); i++ {
				assert.Contains(t, buf.String(), tt.wantStrings[i], fmt.Sprintf("%s is expected at log out", tt.wantStrings[i]))
			}
		})
	}
}
