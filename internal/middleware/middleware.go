package middleware

import (
	"compress/gzip"
	"go.uber.org/zap"
	"net/http"
)

func Decompress(l *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			if r.Header.Get(`Content-Encoding`) == `gzip` {
				gz, err := gzip.NewReader(r.Body)
				if err == nil {
					r.Body = gz
					err = gz.Close()
				}
				if err != nil {
					l.Error("gzip", zap.Error(err))
				}
			}
			next.ServeHTTP(rw, r)
		})
	}
}

func JSONHeader() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			rw.Header().Set("Content-Type", "application/json; charset=utf-8")
			next.ServeHTTP(rw, r)
		})
	}
}

func TextHeader() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
			next.ServeHTTP(rw, r)
		})
	}
}
