package middleware

import "net/http"

// BodyLimit limits request body size to prevent abuse.
func BodyLimit(maxBytes int64) func(next http.Handler) http.Handler {
	if maxBytes <= 0 {
		maxBytes = 1 << 20 // 1 MiB
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}
