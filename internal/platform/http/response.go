package http

import (
	"net/http"

	"github.com/vaaxooo/xbackend/internal/platform/httputil"
)

// ErrorBody is a stable API error contract.
//
// Example:
//
//	{
//	  "error": {
//	    "code": "unauthorized",
//	    "message": "Unauthorized"
//	  }
//	}
type ErrorBody = httputil.ErrorBody

type ErrorPayload = httputil.ErrorPayload

func WriteJSON(w http.ResponseWriter, status int, v any) {
	httputil.WriteJSON(w, status, v)
}

func WriteError(w http.ResponseWriter, status int, code string, message string) {
	httputil.WriteError(w, status, code, message)
}
