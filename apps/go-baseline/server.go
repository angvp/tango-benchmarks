package main

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/angvp/tango-benchmarks/apps/go-baseline/store"
)

type server struct {
	q      *store.Queries
	tokens *tokens
}

func newServer(q *store.Queries, t *tokens) *server { return &server{q: q, tokens: t} }

func (s *server) routes() http.Handler {
	r := chi.NewRouter()
	r.Post("/api/login", s.login)
	r.Get("/api/pets", s.listPets)
	r.Get("/api/pets/{id}", s.getPet)
	r.Group(func(r chi.Router) {
		r.Use(s.requireToken)
		r.Post("/api/pets", s.createPet)
		r.Put("/api/pets/{id}", s.updatePet)
		r.Post("/api/orders", s.createOrder)
		r.Get("/api/orders/{id}", s.getOrder)
	})
	return r
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
