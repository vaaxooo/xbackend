package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/vaaxooo/xbackend/internal/platform/httputil"
)

// RateLimit is a simple in-memory fixed-window rate limiter.
// It is suitable for single-instance deployments.
// For multi-instance, move state to Redis or an API gateway.
func RateLimit(maxRequests int, window time.Duration) func(next http.Handler) http.Handler {
	if maxRequests <= 0 {
		maxRequests = 10
	}
	if window <= 0 {
		window = time.Minute
	}

	type bucket struct {
		resetAt time.Time
		count   int
	}

	var (
		mu      sync.Mutex
		buckets = make(map[string]*bucket)
	)

	cleanup := func(now time.Time) {
		for k, b := range buckets {
			if now.After(b.resetAt.Add(window)) {
				delete(buckets, k)
			}
		}
	}

	keyFor := func(r *http.Request) string {
		ip := r.RemoteAddr
		if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
			ip = host
		}
		// include path to avoid one endpoint starving others
		return ip + "|" + r.URL.Path
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			now := time.Now().UTC()

			mu.Lock()
			cleanup(now)
			key := keyFor(r)
			b, ok := buckets[key]
			if !ok || now.After(b.resetAt) {
				b = &bucket{resetAt: now.Add(window), count: 0}
				buckets[key] = b
			}
			b.count++
			allowed := b.count <= maxRequests
			resetAt := b.resetAt
			mu.Unlock()

			if !allowed {
				// Standard-ish headers
				w.Header().Set("Retry-After", time.Until(resetAt).Round(time.Second).String())
				httputil.WriteError(w, http.StatusTooManyRequests, "rate_limited", "Too many requests")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
