package middleware

import (
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"

	plog "github.com/vaaxooo/xbackend/internal/platform/log"
)

func AccessLog(logger plog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			dur := time.Since(start)

			// RemoteAddr may contain port; split it to log clean IP.
			remoteIP := r.RemoteAddr
			if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
				remoteIP = host
			}

			// Keep original X-Forwarded-For for debugging proxy chains.
			xff := r.Header.Get("X-Forwarded-For")

			logger.Info(r.Context(), "http request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.Status(),
				"bytes", ww.BytesWritten(),
				"duration_ms", dur.Milliseconds(),
				"request_id", middleware.GetReqID(r.Context()),
				"remote_ip", remoteIP,
				"x_forwarded_for", xff,
			)
		})
	}
}
