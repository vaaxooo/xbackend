package middleware

import (
	"net/http"
	"strings"

	"github.com/vaaxooo/xbackend/internal/modules/users/infrastructure/transport/httpctx"
	phttp "github.com/vaaxooo/xbackend/internal/platform/http"
)

type TokenParser interface {
	Parse(token string) (string, error)
}

func RequireJWT(tp TokenParser) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
				phttp.WriteError(w, http.StatusUnauthorized, "unauthorized", "Unauthorized")
				return
			}

			token := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
			uid, err := tp.Parse(token)
			if err != nil {
				phttp.WriteError(w, http.StatusUnauthorized, "unauthorized", "Unauthorized")
				return
			}

			ctx := httpctx.WithUserID(r.Context(), uid)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
