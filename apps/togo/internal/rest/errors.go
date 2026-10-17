package rest

import (
	"github.com/danielgtaylor/huma/v2"
)

// apiError is the contract's error body, {"error": "<message>"}, with the
// status Huma should answer with.
type apiError struct {
	Status  int    `json:"-"`
	Message string `json:"error"`
}

func (e *apiError) Error() string  { return e.Message }
func (e *apiError) GetStatus() int { return e.Status }

func failure(status int, msg string) error { return &apiError{Status: status, Message: msg} }

func init() {
	// Huma's own errors (a body that is not JSON, a failed schema check) use
	// problem+json by default; the contract wants {"error": ...} everywhere.
	huma.NewError = func(status int, msg string, errs ...error) huma.StatusError {
		return &apiError{Status: status, Message: msg}
	}
}
