package logger

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

func Logger(l *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			t1 := time.Now()
			defer func() {
				scheme := "http"
				if r.TLS != nil {
					scheme = "https"
				}
				l.Info("Served",
					zap.Int("status", ww.Status()),
					zap.String("method", r.Method),
					zap.String("URI", fmt.Sprintf("%s://%s%s %s", scheme, r.Host, r.RequestURI, r.Proto)),
					zap.Int("size", ww.BytesWritten()),
					zap.Duration("time", time.Since(t1)),
					zap.String("from", r.RemoteAddr),
				)
			}()
			next.ServeHTTP(ww, r)
		})
	}
}
