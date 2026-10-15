package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/angvp/tango-benchmarks/apps/go-baseline/store"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
	maxInt32        = 1<<31 - 1
)

type petJSON struct {
	ID         int32  `json:"id"`
	Name       string `json:"name"`
	PriceCents int32  `json:"price_cents"`
	Status     string `json:"status"`
	CategoryID int32  `json:"category_id"`
}

func toJSON(p store.Pet) petJSON {
	return petJSON{p.ID, p.Name, p.PriceCents, p.Status, p.CategoryID}
}

func validStatus(s string) bool { return s == "available" || s == "pending" || s == "sold" }

// intParam reads an optional integer query parameter.
func intParam(r *http.Request, name string) (pgtype.Int4, bool) {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		return pgtype.Int4{}, true
	}
	n, err := strconv.ParseInt(raw, 10, 32)
	if err != nil {
		return pgtype.Int4{}, false
	}
	return pgtype.Int4{Int32: int32(n), Valid: true}, true
}

func (s *server) listPets(w http.ResponseWriter, r *http.Request) {
	category, ok1 := intParam(r, "category")
	minPrice, ok2 := intParam(r, "min_price")
	maxPrice, ok3 := intParam(r, "max_price")
	page, ok4 := intParam(r, "page")
	size, ok5 := intParam(r, "page_size")
	if !(ok1 && ok2 && ok3 && ok4 && ok5) {
		writeError(w, http.StatusBadRequest, "numeric parameters must be integers")
		return
	}
	pageNo, pageSize := int32(1), int32(defaultPageSize)
	if page.Valid {
		pageNo = page.Int32
	}
	if size.Valid {
		pageSize = size.Int32
	}
	if pageNo < 1 || pageSize < 1 || pageSize > maxPageSize {
		writeError(w, http.StatusBadRequest, "page must be at least 1 and page_size between 1 and 100")
		return
	}
	var status pgtype.Text
	if v := r.URL.Query().Get("status"); v != "" {
		if !validStatus(v) {
			writeError(w, http.StatusBadRequest, "unknown status")
			return
		}
		status = pgtype.Text{String: v, Valid: true}
	}
	total, err := s.q.CountPets(r.Context(), store.CountPetsParams{Category: category, Status: status, MinPrice: minPrice, MaxPrice: maxPrice})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	rows, err := s.q.ListPets(r.Context(), store.ListPetsParams{
		Category: category, Status: status, MinPrice: minPrice, MaxPrice: maxPrice,
		Skip: (pageNo - 1) * pageSize, PageSize: pageSize,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	items := make([]petJSON, 0, len(rows))
	for _, p := range rows {
		items = append(items, toJSON(p))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "page": pageNo, "page_size": pageSize, "total": total})
}

func idParam(r *http.Request) (int32, bool) {
	n, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 32)
	return int32(n), err == nil && n > 0
}

func (s *server) getPet(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r)
	if !ok {
		writeError(w, http.StatusNotFound, "no such pet")
		return
	}
	p, err := s.q.GetPet(r.Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "no such pet")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	writeJSON(w, http.StatusOK, toJSON(p))
}

type petInput struct {
	Name       string `json:"name"`
	PriceCents int64  `json:"price_cents"`
	Status     string `json:"status"`
	CategoryID int64  `json:"category_id"`
}

// check reports what is wrong with the input, or "" when it is valid.
func (in petInput) check(s *server, r *http.Request) (string, error) {
	switch {
	case in.Name == "":
		return "name is required", nil
	case in.PriceCents < 0 || in.PriceCents > maxInt32:
		return "price_cents must be between 0 and 2147483647", nil
	case !validStatus(in.Status):
		return "status must be available, pending or sold", nil
	case in.CategoryID < 1 || in.CategoryID > maxInt32:
		return "category_id names no category", nil
	}
	exists, err := s.q.CategoryExists(r.Context(), int32(in.CategoryID))
	if err != nil {
		return "", err
	}
	if !exists {
		return "category_id names no category", nil
	}
	return "", nil
}

func (s *server) createPet(w http.ResponseWriter, r *http.Request) {
	var in petInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "body is not valid JSON")
		return
	}
	problem, err := in.check(s, r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	if problem != "" {
		writeError(w, http.StatusUnprocessableEntity, problem)
		return
	}
	p, err := s.q.CreatePet(r.Context(), store.CreatePetParams{Name: in.Name, PriceCents: int32(in.PriceCents), Status: in.Status, CategoryID: int32(in.CategoryID)})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	writeJSON(w, http.StatusCreated, toJSON(p))
}

func (s *server) updatePet(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r)
	if !ok {
		writeError(w, http.StatusNotFound, "no such pet")
		return
	}
	if _, err := s.q.GetPet(r.Context(), id); errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "no such pet")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	var in petInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "body is not valid JSON")
		return
	}
	problem, err := in.check(s, r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	if problem != "" {
		writeError(w, http.StatusUnprocessableEntity, problem)
		return
	}
	p, err := s.q.UpdatePet(r.Context(), store.UpdatePetParams{ID: id, Name: in.Name, PriceCents: int32(in.PriceCents), Status: in.Status, CategoryID: int32(in.CategoryID)})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	writeJSON(w, http.StatusOK, toJSON(p))
}
