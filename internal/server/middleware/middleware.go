package middleware

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/MrSwed/go-musthave-metrics/internal/server/config"
	"go.uber.org/zap"
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

func CheckSign(conf *config.WEB, l *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			if conf != nil && conf.Key != "" && r.Header.Get("HashSHA256") != "" {
				getSha, err := hex.DecodeString(r.Header.Get("HashSHA256"))
				if len(getSha) == 0 || err != nil {
					rw.WriteHeader(http.StatusBadRequest)
					if _, err = rw.Write([]byte("Bad HashKey")); err != nil {
						l.Error("Error return answer", zap.Error(err))
					}
					return
				}
				h := hmac.New(sha256.New, []byte(conf.Key))
				body, err := io.ReadAll(r.Body)
				if err != nil {
					rw.WriteHeader(http.StatusInternalServerError)
					l.Error(err.Error())
					return
				}
				r.Body = io.NopCloser(bytes.NewReader(body))
				h.Write(body)
				if !bytes.Equal(getSha, h.Sum(nil)) {
					rw.WriteHeader(http.StatusBadRequest)
					if _, err = rw.Write([]byte("Bad HashKey")); err != nil {
						l.Error("Error return answer", zap.Error(err))
					}
					return
				}
			}
			next.ServeHTTP(rw, r)
		})
	}
}
