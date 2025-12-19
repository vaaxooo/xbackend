package users

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/vaaxooo/xbackend/internal/platform/http/users/dto"
	"github.com/vaaxooo/xbackend/internal/platform/httputil"
	"github.com/vaaxooo/xbackend/internal/test/contracts"
)

func TestDTOJsonSchemasMatchStructs(t *testing.T) {
	cases := []struct {
		name  string
		value any
		file  string
	}{
		{"register request", dto.RegisterRequest{}, schemaPath("register_request.schema.json")},
		{"login request", dto.LoginRequest{}, schemaPath("login_request.schema.json")},
		{"telegram login request", dto.TelegramLoginRequest{}, schemaPath("telegram_login_request.schema.json")},
		{"refresh request", dto.RefreshRequest{}, schemaPath("refresh_request.schema.json")},
		{"update profile request", dto.UpdateProfileRequest{}, schemaPath("update_profile_request.schema.json")},
		{"link provider request", dto.LinkProviderRequest{}, schemaPath("link_provider_request.schema.json")},
		{"user profile response", dto.UserProfileResponse{}, schemaPath("user_profile_response.schema.json")},
		{"tokens response", dto.TokensResponse{}, schemaPath("tokens_response.schema.json")},
		{"login response", dto.LoginResponse{}, schemaPath("login_response.schema.json")},
		{"link provider response", dto.LinkProviderResponse{}, schemaPath("link_provider_response.schema.json")},
		{"error response", httputil.ErrorBody{}, schemaPath("error_response.schema.json")},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			schema, err := contracts.BuildSchemaForStruct(tt.value)
			if err != nil {
				t.Fatalf("build schema: %v", err)
			}
			canonical, err := contracts.Canonicalize(schema)
			if err != nil {
				t.Fatalf("canonicalize: %v", err)
			}

			raw, err := os.ReadFile(tt.file)
			if err != nil {
				t.Fatalf("read schema file: %v", err)
			}
			fileSchema, err := contracts.LoadSchema(raw)
			if err != nil {
				t.Fatalf("decode file schema: %v", err)
			}
			fileCanonical, err := contracts.Canonicalize(fileSchema)
			if err != nil {
				t.Fatalf("canonicalize file: %v", err)
			}

			if string(fileCanonical) != string(canonical) {
				t.Fatalf("schema mismatch for %s\nexpected:\n%s\nactual:\n%s", tt.file, string(canonical), string(fileCanonical))
			}
		})
	}
}

func schemaPath(name string) string {
	return filepath.Join("schemas", name)
}
