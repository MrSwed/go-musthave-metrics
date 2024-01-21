package middleware

import (
	"compress/gzip"
	"net/http"
)

func Decompress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Header.Get(`Content-Encoding`) == `gzip` {
			if gz, err := gzip.NewReader(r.Body); err == nil {
				r.Body = gz
				_ = gz.Close()
			}
		}
		next.ServeHTTP(rw, r)
	})
}
