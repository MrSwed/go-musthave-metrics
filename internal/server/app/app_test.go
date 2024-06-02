package app

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/MrSwed/go-musthave-metrics/internal/server/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

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
	}{
		{
			name: "Server app run. default",
			fields: func() fields {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				return fields{
					ctx:  ctx,
					stop: cancel,
					cfg: func() *config.Config {
						cfg := config.NewConfig()
						cfg.FileStoragePath = filepath.Join(t.TempDir(), fmt.Sprintf("metrict-db-%d.json", rand.Int()))
						return cfg
					}(),
					build: BuildMetadata{
						Version: "1.0-testing",
						Date:    "24.05.24",
						Commit:  "444333",
					},
				}
			}(),
			wantStrings: []string{
				`"Init app"`,
				`"Build version":"1.0-testing"`,
				`"Build date":"24.05.24"`,
				`"Build commit":"444333"`,
				`Start server`,
				`Server started`,
				`Shutting down server gracefully`,
				`Store save on interval finished`,
				`Storage saved`,
				`Server stopped`,
			},
		},
		{
			name: "Server app run. port busy",
			fields: func() fields {
				cfg := config.NewConfig()
				cfg.FileStoragePath = ""

				portUse, err := net.Listen("tcp", cfg.Address)
				require.NoError(t, err)

				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				return fields{
					ctx: ctx,
					stop: func() {
						_ = portUse.Close()
						cancel()
					},
					cfg: cfg,
					build: BuildMetadata{
						Version: "1.0-testing",
						Date:    "24.05.24",
						Commit:  "444333",
					},
				}
			}(),
			wantStrings: []string{
				`"Init app"`,
				`"Build version":"1.0-testing"`,
				`"Build date":"24.05.24"`,
				`"Build commit":"444333"`,
				`Start server`,
				`Server started`,
				`Shutting down server gracefully`,
				`Store save on interval finished`,
				`Storage saved`,
				`Server stopped`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, stop := signal.NotifyContext(tt.fields.ctx, os.Interrupt, syscall.SIGTERM)
			defer stop()
			defer tt.fields.stop()

			_, err := tt.fields.cfg.Init()
			require.NoError(t, err)

			var buf bytes.Buffer
			logger := zap.New(func(pipeTo io.Writer) zapcore.Core {
				return zapcore.NewCore(
					zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
					zap.CombineWriteSyncers(os.Stderr, zapcore.AddSync(pipeTo)),
					zapcore.InfoLevel,
				)
			}(&buf))

			appHandler := NewApp(ctx, tt.fields.cfg, tt.fields.build, logger)

			appHandler.Run()
			appHandler.Stop()

			t.Log(buf.String())
			for i := 0; i < len(tt.wantStrings); i++ {
				assert.Contains(t, buf.String(), tt.wantStrings[i], fmt.Sprintf("%s is expected at log out", tt.wantStrings[i]))
			}
		})
	}
}
