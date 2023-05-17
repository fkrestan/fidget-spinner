package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

func Logging(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sw := newStatusWriter(w)
			t := time.Now()
			level := zap.InfoLevel
			defer func() {
				if sw.statusCode > 499 {
					level = zap.ErrorLevel
				}

				logger.Log(
					level, "",
					zap.String("method", r.Method),
					zap.String("url", r.URL.Path),
					zap.String("proto", r.Proto),
					zap.Int("status_code", sw.statusCode),
					zap.Duration("duration_ns", time.Since(t)),
					zap.String("remote_addr", r.RemoteAddr),
					zap.String("ua", r.UserAgent()),
				)
			}()
			next.ServeHTTP(sw, r)
		})
	}
}
