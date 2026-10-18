package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/angvp/tango-benchmarks/apps/gin/store"
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
func intParam(c *gin.Context, name string) (pgtype.Int4, bool) {
	raw := c.Query(name)
	if raw == "" {
		return pgtype.Int4{}, true
	}
	n, err := strconv.ParseInt(raw, 10, 32)
	if err != nil {
		return pgtype.Int4{}, false
	}
	return pgtype.Int4{Int32: int32(n), Valid: true}, true
}

func (s *server) listPets(c *gin.Context) {
	category, ok1 := intParam(c, "category")
	minPrice, ok2 := intParam(c, "min_price")
	maxPrice, ok3 := intParam(c, "max_price")
	page, ok4 := intParam(c, "page")
	size, ok5 := intParam(c, "page_size")
	if !(ok1 && ok2 && ok3 && ok4 && ok5) {
		writeError(c, http.StatusBadRequest, "numeric parameters must be integers")
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
		writeError(c, http.StatusBadRequest, "page must be at least 1 and page_size between 1 and 100")
		return
	}
	var status pgtype.Text
	if v := c.Query("status"); v != "" {
		if !validStatus(v) {
			writeError(c, http.StatusBadRequest, "unknown status")
			return
		}
		status = pgtype.Text{String: v, Valid: true}
	}
	total, err := s.q.CountPets(c.Request.Context(), store.CountPetsParams{Category: category, Status: status, MinPrice: minPrice, MaxPrice: maxPrice})
	if err != nil {
		writeError(c, http.StatusInternalServerError, "database error")
		return
	}
	rows, err := s.q.ListPets(c.Request.Context(), store.ListPetsParams{
		Category: category, Status: status, MinPrice: minPrice, MaxPrice: maxPrice,
		Skip: (pageNo - 1) * pageSize, PageSize: pageSize,
	})
	if err != nil {
		writeError(c, http.StatusInternalServerError, "database error")
		return
	}
	items := make([]petJSON, 0, len(rows))
	for _, p := range rows {
		items = append(items, toJSON(p))
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "page": pageNo, "page_size": pageSize, "total": total})
}

func idParam(c *gin.Context) (int32, bool) {
	n, err := strconv.ParseInt(c.Param("id"), 10, 32)
	return int32(n), err == nil && n > 0
}

func (s *server) getPet(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		writeError(c, http.StatusNotFound, "no such pet")
		return
	}
	p, err := s.q.GetPet(c.Request.Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(c, http.StatusNotFound, "no such pet")
		return
	}
	if err != nil {
		writeError(c, http.StatusInternalServerError, "database error")
		return
	}
	c.JSON(http.StatusOK, toJSON(p))
}

type petInput struct {
	Name       string `json:"name"`
	PriceCents int64  `json:"price_cents"`
	Status     string `json:"status"`
	CategoryID int64  `json:"category_id"`
}

// check reports what is wrong with the input, or "" when it is valid.
func (in petInput) check(s *server, c *gin.Context) (string, error) {
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
	exists, err := s.q.CategoryExists(c.Request.Context(), int32(in.CategoryID))
	if err != nil {
		return "", err
	}
	if !exists {
		return "category_id names no category", nil
	}
	return "", nil
}

func (s *server) createPet(c *gin.Context) {
	var in petInput
	if err := c.ShouldBindJSON(&in); err != nil {
		writeError(c, http.StatusBadRequest, "body is not valid JSON")
		return
	}
	problem, err := in.check(s, c)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "database error")
		return
	}
	if problem != "" {
		writeError(c, http.StatusUnprocessableEntity, problem)
		return
	}
	p, err := s.q.CreatePet(c.Request.Context(), store.CreatePetParams{Name: in.Name, PriceCents: int32(in.PriceCents), Status: in.Status, CategoryID: int32(in.CategoryID)})
	if err != nil {
		writeError(c, http.StatusInternalServerError, "database error")
		return
	}
	c.JSON(http.StatusCreated, toJSON(p))
}

func (s *server) updatePet(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		writeError(c, http.StatusNotFound, "no such pet")
		return
	}
	if _, err := s.q.GetPet(c.Request.Context(), id); errors.Is(err, pgx.ErrNoRows) {
		writeError(c, http.StatusNotFound, "no such pet")
		return
	} else if err != nil {
		writeError(c, http.StatusInternalServerError, "database error")
		return
	}
	var in petInput
	if err := c.ShouldBindJSON(&in); err != nil {
		writeError(c, http.StatusBadRequest, "body is not valid JSON")
		return
	}
	problem, err := in.check(s, c)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "database error")
		return
	}
	if problem != "" {
		writeError(c, http.StatusUnprocessableEntity, problem)
		return
	}
	p, err := s.q.UpdatePet(c.Request.Context(), store.UpdatePetParams{ID: id, Name: in.Name, PriceCents: int32(in.PriceCents), Status: in.Status, CategoryID: int32(in.CategoryID)})
	if err != nil {
		writeError(c, http.StatusInternalServerError, "database error")
		return
	}
	c.JSON(http.StatusOK, toJSON(p))
}
