package app

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/MrSwed/go-musthave-metrics/internal/agent/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_app_Run(t *testing.T) {
	type fields struct {
		ctx   context.Context
		stop  context.CancelFunc
		cfg   *config.Config
		build BuildMetadata
	}
	tests := []struct {
		name             string
		fields           fields
		wantStrings      []string
		doNotWantStrings []string
	}{{
		name: "Test_Run",
		fields: func() fields {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			return fields{
				ctx:  ctx,
				stop: cancel,
				cfg:  config.NewConfig(),
				build: BuildMetadata{
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
			`Report interval: 10`,
			`Poll interval: 2`,
			`Rate limit: 1`,
			`Number of metrics at once: 10`,
			`Metric names count: 32`,
			`daemon started: collect runtime metrics with interval 2`,
			`daemon started: collect psutil metrics with interval 2`,
			`daemon started: send metrics interval 10`,
			`Collect runtime metrics`,
			`Collect psutil metrics`,
			`PSUtil metrics collector is stopped`,
			`Runtime metrics collector is stopped`,
			`Metrics sender is stopped`,
			`Agent stopped`,
		},
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, stop := signal.NotifyContext(tt.fields.ctx, os.Interrupt, syscall.SIGTERM)
			defer stop()
			defer tt.fields.stop()

			require.NoError(t, tt.fields.cfg.Init())

			appHandler := NewApp(ctx, tt.fields.cfg, tt.fields.build)

			var buf bytes.Buffer
			log.SetOutput(&buf)
			defer func() {
				log.SetOutput(os.Stderr)
			}()

			appHandler.Run()

			t.Log(buf.String())
			for i := 0; i < len(tt.wantStrings); i++ {
				assert.Contains(t, buf.String(), tt.wantStrings[i], fmt.Sprintf("%s is expected at log out", tt.wantStrings[i]))
			}
		})
	}
}
