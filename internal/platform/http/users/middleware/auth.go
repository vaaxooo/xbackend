package middleware

import (
	"net/http"
	"strings"

	"github.com/vaaxooo/xbackend/internal/modules/users/public"
	phttp "github.com/vaaxooo/xbackend/internal/platform/http"
	"github.com/vaaxooo/xbackend/internal/platform/http/users/httpctx"
)

func RequireJWT(auth public.AuthPort) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				phttp.WriteError(w, http.StatusUnauthorized, "unauthorized", "Unauthorized")
				return
			}

			token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
			ctx, err := auth.Verify(token)
			if err != nil {
				phttp.WriteError(w, http.StatusUnauthorized, "unauthorized", "Unauthorized")
				return
			}

			reqCtx := httpctx.WithUserID(r.Context(), ctx.UserID)
			next.ServeHTTP(w, r.WithContext(reqCtx))
		})
	}
}
