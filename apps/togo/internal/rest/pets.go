package rest

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
	maxInt32        = 1<<31 - 1
)

type petOutput struct {
	ID         int32  `json:"id"`
	Name       string `json:"name"`
	PriceCents int32  `json:"price_cents"`
	Status     string `json:"status"`
	CategoryID int32  `json:"category_id"`
}

type petBody struct {
	Name       string `json:"name"`
	PriceCents int64  `json:"price_cents"`
	Status     string `json:"status"`
	CategoryID int64  `json:"category_id"`
}

type petResponse struct{ Body petOutput }

type petInput struct{ Body petBody }

type petUpdateInput struct {
	ID   string `path:"id"`
	Body petBody
}

type idInput struct {
	ID string `path:"id"`
}

type listPetsInput struct {
	Category string `query:"category"`
	Status   string `query:"status"`
	MinPrice string `query:"min_price"`
	MaxPrice string `query:"max_price"`
	Page     string `query:"page"`
	PageSize string `query:"page_size"`
}

type listPetsOutput struct {
	Body struct {
		Items    []petOutput `json:"items"`
		Page     int32       `json:"page"`
		PageSize int32       `json:"page_size"`
		Total    int64       `json:"total"`
	}
}

func validStatus(s string) bool { return s == "available" || s == "pending" || s == "sold" }

// optionalInt parses an optional integer parameter; a missing one is NULL.
func optionalInt(raw string) (sql.NullInt32, bool) {
	if raw == "" {
		return sql.NullInt32{}, true
	}
	n, err := strconv.ParseInt(raw, 10, 32)
	if err != nil {
		return sql.NullInt32{}, false
	}
	return sql.NullInt32{Int32: int32(n), Valid: true}, true
}

func idOf(raw string) (int32, bool) {
	n, err := strconv.ParseInt(raw, 10, 32)
	return int32(n), err == nil && n > 0
}

const petFilter = `WHERE ($1::int IS NULL OR category_id = $1)
  AND ($2::text IS NULL OR status = $2)
  AND ($3::int IS NULL OR price_cents >= $3)
  AND ($4::int IS NULL OR price_cents <= $4)`

func (s *shop) listPets(ctx context.Context, in *listPetsInput) (*listPetsOutput, error) {
	category, ok1 := optionalInt(in.Category)
	minPrice, ok2 := optionalInt(in.MinPrice)
	maxPrice, ok3 := optionalInt(in.MaxPrice)
	page, ok4 := optionalInt(in.Page)
	size, ok5 := optionalInt(in.PageSize)
	if !(ok1 && ok2 && ok3 && ok4 && ok5) {
		return nil, failure(http.StatusBadRequest, "numeric parameters must be integers")
	}
	pageNo, pageSize := int32(1), int32(defaultPageSize)
	if page.Valid {
		pageNo = page.Int32
	}
	if size.Valid {
		pageSize = size.Int32
	}
	if pageNo < 1 || pageSize < 1 || pageSize > maxPageSize {
		return nil, failure(http.StatusBadRequest, "page must be at least 1 and page_size between 1 and 100")
	}
	status := sql.NullString{}
	if in.Status != "" {
		if !validStatus(in.Status) {
			return nil, failure(http.StatusBadRequest, "unknown status")
		}
		status = sql.NullString{String: in.Status, Valid: true}
	}
	out := &listPetsOutput{}
	out.Body.Page, out.Body.PageSize = pageNo, pageSize
	if err := s.a.SQLDB.QueryRowContext(ctx, `SELECT count(*) FROM pets `+petFilter, category, status, minPrice, maxPrice).Scan(&out.Body.Total); err != nil {
		return nil, err
	}
	rows, err := s.a.SQLDB.QueryContext(ctx, `SELECT id, name, price_cents, status, category_id FROM pets `+petFilter+
		` ORDER BY id LIMIT $5 OFFSET $6`, category, status, minPrice, maxPrice, pageSize, (pageNo-1)*pageSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out.Body.Items = []petOutput{}
	for rows.Next() {
		var p petOutput
		if err := rows.Scan(&p.ID, &p.Name, &p.PriceCents, &p.Status, &p.CategoryID); err != nil {
			return nil, err
		}
		out.Body.Items = append(out.Body.Items, p)
	}
	return out, rows.Err()
}

func (s *shop) petByID(ctx context.Context, id int32) (*petOutput, error) {
	var p petOutput
	err := s.a.SQLDB.QueryRowContext(ctx, `SELECT id, name, price_cents, status, category_id FROM pets WHERE id = $1`, id).
		Scan(&p.ID, &p.Name, &p.PriceCents, &p.Status, &p.CategoryID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, failure(http.StatusNotFound, "no such pet")
	}
	return &p, err
}

func (s *shop) getPet(ctx context.Context, in *idInput) (*petResponse, error) {
	id, ok := idOf(in.ID)
	if !ok {
		return nil, failure(http.StatusNotFound, "no such pet")
	}
	p, err := s.petByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &petResponse{Body: *p}, nil
}

// check returns the 422 error for an invalid body, or nil.
func (s *shop) check(ctx context.Context, in petBody) error {
	unprocessable := func(msg string) error { return failure(http.StatusUnprocessableEntity, msg) }
	switch {
	case strings.TrimSpace(in.Name) == "":
		return unprocessable("name is required")
	case in.PriceCents < 0 || in.PriceCents > maxInt32:
		return unprocessable("price_cents must be between 0 and 2147483647")
	case !validStatus(in.Status):
		return unprocessable("status must be available, pending or sold")
	case in.CategoryID < 1 || in.CategoryID > maxInt32:
		return unprocessable("category_id names no category")
	}
	var exists bool
	if err := s.a.SQLDB.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM categories WHERE id = $1)`, in.CategoryID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return unprocessable("category_id names no category")
	}
	return nil
}

func (s *shop) createPet(ctx context.Context, in *petInput) (*petResponse, error) {
	if err := s.check(ctx, in.Body); err != nil {
		return nil, err
	}
	var p petOutput
	err := s.a.SQLDB.QueryRowContext(ctx, `INSERT INTO pets (name, price_cents, status, category_id) VALUES ($1, $2, $3, $4)
		RETURNING id, name, price_cents, status, category_id`, in.Body.Name, in.Body.PriceCents, in.Body.Status, in.Body.CategoryID).
		Scan(&p.ID, &p.Name, &p.PriceCents, &p.Status, &p.CategoryID)
	return &petResponse{Body: p}, err
}

func (s *shop) updatePet(ctx context.Context, in *petUpdateInput) (*petResponse, error) {
	id, ok := idOf(in.ID)
	if !ok {
		return nil, failure(http.StatusNotFound, "no such pet")
	}
	if _, err := s.petByID(ctx, id); err != nil {
		return nil, err
	}
	if err := s.check(ctx, in.Body); err != nil {
		return nil, err
	}
	var p petOutput
	err := s.a.SQLDB.QueryRowContext(ctx, `UPDATE pets SET name = $2, price_cents = $3, status = $4, category_id = $5 WHERE id = $1
		RETURNING id, name, price_cents, status, category_id`, id, in.Body.Name, in.Body.PriceCents, in.Body.Status, in.Body.CategoryID).
		Scan(&p.ID, &p.Name, &p.PriceCents, &p.Status, &p.CategoryID)
	return &petResponse{Body: p}, err
}
