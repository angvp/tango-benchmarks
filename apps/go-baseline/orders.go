package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"

	"github.com/angvp/tango-benchmarks/apps/go-baseline/store"
)

type orderJSON struct {
	ID         int32 `json:"id"`
	PetID      int32 `json:"pet_id"`
	CustomerID int32 `json:"customer_id"`
}

func (s *server) createOrder(w http.ResponseWriter, r *http.Request) {
	var in struct {
		PetID      int64 `json:"pet_id"`
		CustomerID int64 `json:"customer_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "body is not valid JSON")
		return
	}
	if in.PetID < 1 || in.PetID > maxInt32 || in.CustomerID < 1 || in.CustomerID > maxInt32 {
		writeError(w, http.StatusUnprocessableEntity, "pet_id and customer_id must name existing rows")
		return
	}
	petOK, err := s.q.PetExists(r.Context(), int32(in.PetID))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	customerOK, err := s.q.CustomerExists(r.Context(), int32(in.CustomerID))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	if !petOK || !customerOK {
		writeError(w, http.StatusUnprocessableEntity, "pet_id and customer_id must name existing rows")
		return
	}
	o, err := s.q.CreateOrder(r.Context(), store.CreateOrderParams{PetID: int32(in.PetID), CustomerID: int32(in.CustomerID)})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	writeJSON(w, http.StatusCreated, orderJSON{o.ID, o.PetID, o.CustomerID})
}

func (s *server) getOrder(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r)
	if !ok {
		writeError(w, http.StatusNotFound, "no such order")
		return
	}
	o, err := s.q.GetOrder(r.Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "no such order")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	writeJSON(w, http.StatusOK, orderJSON{o.ID, o.PetID, o.CustomerID})
}
