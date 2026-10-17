package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/go-fuego/fuego/option"

	"github.com/angvp/tango-benchmarks/apps/fuego/store"
)

type shop struct {
	q      *store.Queries
	tokens *tokens
}

func newShop(q *store.Queries, t *tokens) *shop { return &shop{q: q, tokens: t} }

func (sh *shop) routes(s *fuego.Server) {
	auth := option.Middleware(sh.requireToken)
	created := option.DefaultStatusCode(http.StatusCreated)

	fuego.Post(s, "/api/login", sh.login)
	fuego.Get(s, "/api/pets", sh.listPets)
	fuego.Get(s, "/api/pets/{id}", sh.getPet)
	fuego.Post(s, "/api/pets", sh.createPet, auth, created)
	fuego.Put(s, "/api/pets/{id}", sh.updatePet, auth)
	fuego.Post(s, "/api/orders", sh.createOrder, auth, created)
	fuego.Get(s, "/api/orders/{id}", sh.getOrder, auth)
}

// apiError is an error with the status and message the contract asks for.
type apiError struct {
	status int
	msg    string
}

func (e apiError) Error() string   { return e.msg }
func (e apiError) StatusCode() int { return e.status }

func failure(status int, msg string) error { return apiError{status, msg} }

// sendError writes every error as {"error": "<message>"}, the contract's shape.
func sendError(w http.ResponseWriter, _ *http.Request, err error) {
	status, msg := http.StatusInternalServerError, "internal error"
	var withStatus fuego.ErrorWithStatus
	if errors.As(err, &withStatus) {
		status, msg = withStatus.StatusCode(), http.StatusText(withStatus.StatusCode())
	}
	var api apiError
	if errors.As(err, &api) {
		msg = api.msg
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
