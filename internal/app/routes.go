package app

import (
	"github.com/go-chi/chi/v5"

	usershttp "github.com/vaaxooo/xbackend/internal/modules/users/transport/http"
)

// RegisterAPIV1 registers all HTTP routes for API v1.
// Versioning is done at the router boundary to keep handlers clean.
func RegisterAPIV1(r chi.Router, modules *Modules) {
	usershttp.RegisterV1(r, modules.Users.Service, modules.Users.Tokens)
}
