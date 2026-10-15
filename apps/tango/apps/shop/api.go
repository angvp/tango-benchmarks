package shop

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/angvp/tango"
	"github.com/angvp/tango/admin"
	"github.com/angvp/tango/auth"
	tangojwt "github.com/angvp/tango/auth/jwt"
	"github.com/angvp/tango/db"
	"github.com/angvp/tango/model"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
	tokenLifetime   = 12 * time.Hour
)

type api struct {
	store            *db.Store
	tokens           *tangojwt.Service
	registry         *tango.Registry
	pet, order, cust model.ModelMeta
}

func newAPI(registry *tango.Registry, store *db.Store, tokens *tangojwt.Service) (*api, error) {
	a := &api{store: store, tokens: tokens, registry: registry}
	for name, dst := range map[string]*model.ModelMeta{"Pet": &a.pet, "Order": &a.order, "Customer": &a.cust} {
		meta, ok := registry.Models().Get(name)
		if !ok {
			return nil, errors.New("model " + name + " is not registered")
		}
		*dst = meta
	}
	return a, nil
}

func fail(ctx *tango.Context, status int, message string) error {
	return ctx.JSON(status, map[string]string{"error": message})
}

func (a *api) login(ctx *tango.Context) error {
	var in struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := ctx.Bind(&in); err != nil {
		return fail(ctx, http.StatusBadRequest, "body is not valid JSON")
	}
	// The admin app registers AdminUser after this app (it must come last in
	// InstalledApps), so the model is looked up when a request needs it.
	adminUser, ok := a.registry.Models().Get("AdminUser")
	if !ok {
		return errors.New("the admin app is not installed")
	}
	var users []admin.AdminUser
	query := db.Query{Where: []db.Condition{{Field: "Username", Op: db.OpEq, Value: in.Username}}, Limit: 1}
	if err := a.store.List(ctx.Context(), adminUser, query, &users); err != nil {
		return err
	}
	if len(users) != 1 || !users[0].Active || !auth.VerifyPassword(users[0].PasswordHash, in.Password) {
		return fail(ctx, http.StatusUnauthorized, "bad credentials")
	}
	token, err := a.tokens.Issue(strconv.FormatInt(users[0].ID, 10), tokenLifetime)
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, map[string]string{"token": token})
}

// intQuery reads an optional integer query parameter.
func intQuery(ctx *tango.Context, name string) (value int64, present bool, ok bool) {
	raw := ctx.Query(name)
	if raw == "" {
		return 0, false, true
	}
	n, err := strconv.ParseInt(raw, 10, 64)
	return n, true, err == nil
}

func validStatus(s string) bool { return s == "available" || s == "pending" || s == "sold" }

func (a *api) listPets(ctx *tango.Context) error {
	category, hasCategory, ok1 := intQuery(ctx, "category")
	minPrice, hasMin, ok2 := intQuery(ctx, "min_price")
	maxPrice, hasMax, ok3 := intQuery(ctx, "max_price")
	page, hasPage, ok4 := intQuery(ctx, "page")
	size, hasSize, ok5 := intQuery(ctx, "page_size")
	if !(ok1 && ok2 && ok3 && ok4 && ok5) {
		return fail(ctx, http.StatusBadRequest, "numeric parameters must be integers")
	}
	if !hasPage {
		page = 1
	}
	if !hasSize {
		size = defaultPageSize
	}
	if page < 1 || size < 1 || size > maxPageSize {
		return fail(ctx, http.StatusBadRequest, "page must be at least 1 and page_size between 1 and 100")
	}
	var where []db.Condition
	if hasCategory {
		where = append(where, db.Condition{Field: "CategoryID", Op: db.OpEq, Value: category})
	}
	if status := ctx.Query("status"); status != "" {
		if !validStatus(status) {
			return fail(ctx, http.StatusBadRequest, "unknown status")
		}
		where = append(where, db.Condition{Field: "Status", Op: db.OpEq, Value: status})
	}
	if hasMin {
		where = append(where, db.Condition{Field: "PriceCents", Op: db.OpGte, Value: minPrice})
	}
	if hasMax {
		where = append(where, db.Condition{Field: "PriceCents", Op: db.OpLte, Value: maxPrice})
	}
	total, err := a.store.Count(ctx.Context(), a.pet, db.Query{Where: where})
	if err != nil {
		return err
	}
	pets := []Pet{}
	query := db.Query{Where: where, OrderBy: []string{"ID"}, Limit: int(size), Offset: int((page - 1) * size)}
	if err := a.store.List(ctx.Context(), a.pet, query, &pets); err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, map[string]any{"items": pets, "page": page, "page_size": size, "total": total})
}

func (a *api) getPet(ctx *tango.Context) error {
	var pet Pet
	if err := a.store.Get(ctx.Context(), a.pet, ctx.Param("id"), &pet); err != nil {
		return notFoundOr(ctx, err, "no such pet")
	}
	return ctx.JSON(http.StatusOK, pet)
}

// notFoundOr answers 404 for a missing row and passes any other error on.
func notFoundOr(ctx *tango.Context, err error, message string) error {
	if errors.Is(err, db.ErrNotFound) {
		return fail(ctx, http.StatusNotFound, message)
	}
	return err
}

// problem names what is wrong with a pet's fields, or "" when they are valid.
func problem(p Pet) string {
	switch {
	case p.Name == "":
		return "name is required"
	case p.PriceCents < 0:
		return "price_cents must not be negative"
	case !validStatus(p.Status):
		return "status must be available, pending or sold"
	}
	return ""
}

func (a *api) writePet(ctx *tango.Context, p *Pet, write func() error, status int) error {
	if msg := problem(*p); msg != "" {
		return fail(ctx, http.StatusUnprocessableEntity, msg)
	}
	if err := write(); err != nil {
		if errors.Is(err, db.ErrInvalidForeignKey) {
			return fail(ctx, http.StatusUnprocessableEntity, "category_id names no category")
		}
		return err
	}
	return ctx.JSON(status, p)
}

func (a *api) createPet(ctx *tango.Context) error {
	var p Pet
	if err := ctx.Bind(&p); err != nil {
		return fail(ctx, http.StatusBadRequest, "body is not valid JSON")
	}
	p.ID = 0
	return a.writePet(ctx, &p, func() error { return a.store.Create(ctx.Context(), a.pet, &p) }, http.StatusCreated)
}

func (a *api) updatePet(ctx *tango.Context) error {
	var current Pet
	if err := a.store.Get(ctx.Context(), a.pet, ctx.Param("id"), &current); err != nil {
		return notFoundOr(ctx, err, "no such pet")
	}
	var p Pet
	if err := ctx.Bind(&p); err != nil {
		return fail(ctx, http.StatusBadRequest, "body is not valid JSON")
	}
	p.ID = current.ID
	return a.writePet(ctx, &p, func() error { return a.store.Update(ctx.Context(), a.pet, &p) }, http.StatusOK)
}

func (a *api) createOrder(ctx *tango.Context) error {
	var o Order
	if err := ctx.Bind(&o); err != nil {
		return fail(ctx, http.StatusBadRequest, "body is not valid JSON")
	}
	o.ID = 0
	if o.PetID < 1 || o.CustomerID < 1 {
		return fail(ctx, http.StatusUnprocessableEntity, "pet_id and customer_id must name existing rows")
	}
	if err := a.store.Create(ctx.Context(), a.order, &o); err != nil {
		if errors.Is(err, db.ErrInvalidForeignKey) {
			return fail(ctx, http.StatusUnprocessableEntity, "pet_id and customer_id must name existing rows")
		}
		return err
	}
	return ctx.JSON(http.StatusCreated, o)
}

func (a *api) getOrder(ctx *tango.Context) error {
	var o Order
	if err := a.store.Get(ctx.Context(), a.order, ctx.Param("id"), &o); err != nil {
		return notFoundOr(ctx, err, "no such order")
	}
	return ctx.JSON(http.StatusOK, o)
}
