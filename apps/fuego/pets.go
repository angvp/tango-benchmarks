package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-fuego/fuego"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/angvp/tango-benchmarks/apps/fuego/store"
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

type petPage struct {
	Items    []petOutput `json:"items"`
	Page     int32       `json:"page"`
	PageSize int32       `json:"page_size"`
	Total    int64       `json:"total"`
}

func toOutput(p store.Pet) petOutput {
	return petOutput{p.ID, p.Name, p.PriceCents, p.Status, p.CategoryID}
}

func validStatus(s string) bool { return s == "available" || s == "pending" || s == "sold" }

// intParam reads an optional integer query parameter.
func intParam(c fuego.ContextNoBody, name string) (pgtype.Int4, bool) {
	raw := c.QueryParam(name)
	if raw == "" {
		return pgtype.Int4{}, true
	}
	n, err := strconv.ParseInt(raw, 10, 32)
	if err != nil {
		return pgtype.Int4{}, false
	}
	return pgtype.Int4{Int32: int32(n), Valid: true}, true
}

func (sh *shop) listPets(c fuego.ContextNoBody) (petPage, error) {
	category, ok1 := intParam(c, "category")
	minPrice, ok2 := intParam(c, "min_price")
	maxPrice, ok3 := intParam(c, "max_price")
	page, ok4 := intParam(c, "page")
	size, ok5 := intParam(c, "page_size")
	if !(ok1 && ok2 && ok3 && ok4 && ok5) {
		return petPage{}, failure(http.StatusBadRequest, "numeric parameters must be integers")
	}
	pageNo, pageSize := int32(1), int32(defaultPageSize)
	if page.Valid {
		pageNo = page.Int32
	}
	if size.Valid {
		pageSize = size.Int32
	}
	if pageNo < 1 || pageSize < 1 || pageSize > maxPageSize {
		return petPage{}, failure(http.StatusBadRequest, "page must be at least 1 and page_size between 1 and 100")
	}
	var status pgtype.Text
	if v := c.QueryParam("status"); v != "" {
		if !validStatus(v) {
			return petPage{}, failure(http.StatusBadRequest, "unknown status")
		}
		status = pgtype.Text{String: v, Valid: true}
	}
	total, err := sh.q.CountPets(c.Context(), store.CountPetsParams{Category: category, Status: status, MinPrice: minPrice, MaxPrice: maxPrice})
	if err != nil {
		return petPage{}, err
	}
	rows, err := sh.q.ListPets(c.Context(), store.ListPetsParams{
		Category: category, Status: status, MinPrice: minPrice, MaxPrice: maxPrice,
		Skip: (pageNo - 1) * pageSize, PageSize: pageSize,
	})
	if err != nil {
		return petPage{}, err
	}
	items := make([]petOutput, 0, len(rows))
	for _, p := range rows {
		items = append(items, toOutput(p))
	}
	return petPage{Items: items, Page: pageNo, PageSize: pageSize, Total: total}, nil
}

// idParam reads the {id} path parameter; ok is false for anything that cannot
// name a row.
func idParam(c interface{ PathParam(string) string }) (int32, bool) {
	n, err := strconv.ParseInt(c.PathParam("id"), 10, 32)
	return int32(n), err == nil && n > 0
}

func (sh *shop) getPet(c fuego.ContextNoBody) (petOutput, error) {
	id, ok := idParam(c)
	if !ok {
		return petOutput{}, failure(http.StatusNotFound, "no such pet")
	}
	p, err := sh.q.GetPet(c.Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		return petOutput{}, failure(http.StatusNotFound, "no such pet")
	}
	if err != nil {
		return petOutput{}, err
	}
	return toOutput(p), nil
}

type petInput struct {
	Name       string `json:"name"`
	PriceCents int64  `json:"price_cents"`
	Status     string `json:"status"`
	CategoryID int64  `json:"category_id"`
}

// check returns the 422 error for invalid input, or nil when it is valid.
func (sh *shop) check(c fuego.ContextWithBody[petInput], in petInput) error {
	switch {
	case in.Name == "":
		return failure(http.StatusUnprocessableEntity, "name is required")
	case in.PriceCents < 0 || in.PriceCents > maxInt32:
		return failure(http.StatusUnprocessableEntity, "price_cents must be between 0 and 2147483647")
	case !validStatus(in.Status):
		return failure(http.StatusUnprocessableEntity, "status must be available, pending or sold")
	case in.CategoryID < 1 || in.CategoryID > maxInt32:
		return failure(http.StatusUnprocessableEntity, "category_id names no category")
	}
	exists, err := sh.q.CategoryExists(c.Context(), int32(in.CategoryID))
	if err != nil {
		return err
	}
	if !exists {
		return failure(http.StatusUnprocessableEntity, "category_id names no category")
	}
	return nil
}

func (sh *shop) createPet(c fuego.ContextWithBody[petInput]) (petOutput, error) {
	in, err := c.Body()
	if err != nil {
		return petOutput{}, err
	}
	if err := sh.check(c, in); err != nil {
		return petOutput{}, err
	}
	p, err := sh.q.CreatePet(c.Context(), store.CreatePetParams{Name: in.Name, PriceCents: int32(in.PriceCents), Status: in.Status, CategoryID: int32(in.CategoryID)})
	if err != nil {
		return petOutput{}, err
	}
	return toOutput(p), nil
}

func (sh *shop) updatePet(c fuego.ContextWithBody[petInput]) (petOutput, error) {
	id, ok := idParam(c)
	if !ok {
		return petOutput{}, failure(http.StatusNotFound, "no such pet")
	}
	if _, err := sh.q.GetPet(c.Context(), id); errors.Is(err, pgx.ErrNoRows) {
		return petOutput{}, failure(http.StatusNotFound, "no such pet")
	} else if err != nil {
		return petOutput{}, err
	}
	in, err := c.Body()
	if err != nil {
		return petOutput{}, err
	}
	if err := sh.check(c, in); err != nil {
		return petOutput{}, err
	}
	p, err := sh.q.UpdatePet(c.Context(), store.UpdatePetParams{ID: id, Name: in.Name, PriceCents: int32(in.PriceCents), Status: in.Status, CategoryID: int32(in.CategoryID)})
	if err != nil {
		return petOutput{}, err
	}
	return toOutput(p), nil
}
