package httputil

import (
	"encoding/json"
	"net/http"
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
type ErrorBody struct {
	Error ErrorPayload `json:"error"`
}

type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type SuccessBody struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

func NewSuccessBody(message string) SuccessBody {
	return SuccessBody{Status: "ok", Message: message}
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func WriteError(w http.ResponseWriter, status int, code string, message string) {
	WriteJSON(w, status, ErrorBody{Error: ErrorPayload{Code: code, Message: message}})
}
