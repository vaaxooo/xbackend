package public_test

import (
	"testing"

	"github.com/vaaxooo/xbackend/internal/modules/users/application/common"
	"github.com/vaaxooo/xbackend/internal/modules/users/infrastructure/auth"
	"github.com/vaaxooo/xbackend/internal/modules/users/infrastructure/tokens"
	"github.com/vaaxooo/xbackend/internal/modules/users/public"
)

func TestAuthAdapterSatisfiesPorts(t *testing.T) {
	var _ public.AuthPort = (*auth.JWTAuth)(nil)
	var _ common.AccessTokenIssuer = (*auth.JWTAuth)(nil)
}

func TestTokenDriverIsNotExposed(t *testing.T) {
	if _, ok := any(&tokens.HS256{}).(public.AuthPort); ok {
		t.Fatalf("token driver should not satisfy public.AuthPort directly")
	}
}
