// Package fakeapp is an in-memory pet shop that follows contract/petshop.md.
// It exists to test the conformance suite: each Fault switches on one
// deliberate deviation, and the suite has to notice it. It is not a benchmark
// subject.
package fakeapp

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/angvp/tango-benchmarks/internal/seed"
)

const (
	Username = "bench-staff"
	Password = "bench-staff-password"
	token    = "fake-token"
)

type Fault string

const (
	WrongTotal                   Fault = "wrong-total"
	UnorderedList                Fault = "unordered-list"
	PageSizeUncapped             Fault = "page-size-uncapped"
	IgnoresPriceFilter           Fault = "ignores-price-filter"
	IgnoresStatusFilter          Fault = "ignores-status-filter"
	NoPastLastPage               Fault = "no-past-last-page"
	MissingPetIs200              Fault = "missing-pet-is-200"
	BadLoginAccepted             Fault = "bad-login-accepted"
	OpenPetCreate                Fault = "open-pet-create"
	OpenPetUpdate                Fault = "open-pet-update"
	OpenOrderCreate              Fault = "open-order-create"
	OpenOrderGet                 Fault = "open-order-get"
	CreateIs200                  Fault = "create-is-200"
	NoPetValidation              Fault = "no-pet-validation"
	UpdateIgnoresBody            Fault = "update-ignores-body"
	UpdateMissingIs200           Fault = "update-missing-is-200"
	NoUpdateValidation           Fault = "no-update-validation"
	OrderUnknownPetAccepted      Fault = "order-unknown-pet-accepted"
	OrderUnknownCustomerAccepted Fault = "order-unknown-customer-accepted"
	OpenAdmin                    Fault = "open-admin"
	AdminListEmpty               Fault = "admin-list-empty"
	ExtraPet                     Fault = "extra-pet"
	NoAdmin                      Fault = "no-admin"
)

type App struct {
	mu     sync.Mutex
	faults []Fault
	pets   []seed.Pet
	orders []order
}

type order struct {
	ID         int `json:"id"`
	PetID      int `json:"pet_id"`
	CustomerID int `json:"customer_id"`
}

type pet struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	PriceCents int    `json:"price_cents"`
	Status     string `json:"status"`
	CategoryID int    `json:"category_id"`
}

func New(faults ...Fault) *App {
	a := &App{faults: faults, pets: seed.Pets()}
	if a.has(ExtraPet) {
		a.pets = append(a.pets, seed.Pet{ID: len(a.pets) + 1, Name: "Extra", PriceCents: 1000, Status: "available", CategoryID: 1})
	}
	return a
}

func (a *App) has(f Fault) bool { return slices.Contains(a.faults, f) }

func (a *App) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/login", a.login)
	mux.HandleFunc("GET /api/pets", a.listPets)
	mux.HandleFunc("GET /api/pets/{id}", a.getPet)
	mux.HandleFunc("POST /api/pets", a.createPet)
	mux.HandleFunc("PUT /api/pets/{id}", a.updatePet)
	mux.HandleFunc("POST /api/orders", a.createOrder)
	mux.HandleFunc("GET /api/orders/{id}", a.getOrder)
	if !a.has(NoAdmin) {
		mux.HandleFunc("GET /admin/login/", a.adminLoginPage)
		mux.HandleFunc("POST /admin/login/", a.adminLogin)
		mux.HandleFunc("GET /admin/petshop/pet/", a.adminList)
	}
	return mux
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func fail(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func (a *App) authed(r *http.Request, open Fault) bool {
	return a.has(open) || r.Header.Get("Authorization") == "Bearer "+token
}

func (a *App) login(w http.ResponseWriter, r *http.Request) {
	var in struct{ Username, Password string }
	_ = json.NewDecoder(r.Body).Decode(&in)
	if in.Username == Username && in.Password == Password || a.has(BadLoginAccepted) {
		writeJSON(w, http.StatusOK, map[string]string{"token": token})
		return
	}
	fail(w, http.StatusUnauthorized, "bad credentials")
}

func (a *App) listPets(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, size, ok := paging(q.Get("page"), q.Get("page_size"), a.has(PageSizeUncapped))
	if !ok {
		fail(w, http.StatusBadRequest, "bad paging")
		return
	}
	status := q.Get("status")
	if status != "" && !validStatus(status) {
		fail(w, http.StatusBadRequest, "bad status")
		return
	}
	category, minPrice, maxPrice := q.Get("category"), q.Get("min_price"), q.Get("max_price")
	for _, v := range []string{category, minPrice, maxPrice} {
		if v != "" {
			if _, err := strconv.Atoi(v); err != nil {
				fail(w, http.StatusBadRequest, "not a number")
				return
			}
		}
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	var matches []pet
	for _, p := range a.pets {
		if category != "" && strconv.Itoa(p.CategoryID) != category {
			continue
		}
		if status != "" && !a.has(IgnoresStatusFilter) && p.Status != status {
			continue
		}
		if !a.has(IgnoresPriceFilter) {
			if v, err := strconv.Atoi(minPrice); err == nil && p.PriceCents < v {
				continue
			}
			if v, err := strconv.Atoi(maxPrice); err == nil && p.PriceCents > v {
				continue
			}
		}
		matches = append(matches, pet(p))
	}
	if a.has(UnorderedList) {
		slices.Reverse(matches)
	}
	total := len(matches)
	if a.has(WrongTotal) && category != "" {
		total++
	}
	start := (page - 1) * size
	var items []pet
	switch {
	case start < len(matches):
		items = matches[start:min(start+size, len(matches))]
	case a.has(NoPastLastPage):
		fail(w, http.StatusNotFound, "no such page")
		return
	default:
		items = []pet{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "page": page, "page_size": size, "total": total})
}

func paging(pageS, sizeS string, uncapped bool) (page, size int, ok bool) {
	page, size = 1, 20
	var err error
	if pageS != "" {
		if page, err = strconv.Atoi(pageS); err != nil || page < 1 {
			return 0, 0, false
		}
	}
	if sizeS != "" {
		if size, err = strconv.Atoi(sizeS); err != nil || size < 1 || (size > 100 && !uncapped) {
			return 0, 0, false
		}
	}
	return page, size, true
}

func validStatus(s string) bool { return s == "available" || s == "pending" || s == "sold" }

func (a *App) find(id int) (int, bool) {
	for i, p := range a.pets {
		if p.ID == id {
			return i, true
		}
	}
	return 0, false
}

func (a *App) getPet(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))
	a.mu.Lock()
	defer a.mu.Unlock()
	i, ok := a.find(id)
	if !ok {
		if a.has(MissingPetIs200) {
			writeJSON(w, http.StatusOK, map[string]any{})
			return
		}
		fail(w, http.StatusNotFound, "no such pet")
		return
	}
	writeJSON(w, http.StatusOK, pet(a.pets[i]))
}

type petInput struct {
	Name       string `json:"name"`
	PriceCents int    `json:"price_cents"`
	Status     string `json:"status"`
	CategoryID int    `json:"category_id"`
}

func (in petInput) valid() bool {
	return in.Name != "" && in.PriceCents >= 0 && validStatus(in.Status) && in.CategoryID >= 1 && in.CategoryID <= seed.CategoryCount
}

func (a *App) createPet(w http.ResponseWriter, r *http.Request) {
	if !a.authed(r, OpenPetCreate) {
		fail(w, http.StatusUnauthorized, "token required")
		return
	}
	var in petInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		fail(w, http.StatusBadRequest, "bad json")
		return
	}
	if !in.valid() && !a.has(NoPetValidation) {
		fail(w, http.StatusUnprocessableEntity, "invalid pet")
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	p := seed.Pet{ID: a.pets[len(a.pets)-1].ID + 1, Name: in.Name, PriceCents: in.PriceCents, Status: in.Status, CategoryID: in.CategoryID}
	a.pets = append(a.pets, p)
	status := http.StatusCreated
	if a.has(CreateIs200) {
		status = http.StatusOK
	}
	writeJSON(w, status, pet(p))
}

func (a *App) updatePet(w http.ResponseWriter, r *http.Request) {
	if !a.authed(r, OpenPetUpdate) {
		fail(w, http.StatusUnauthorized, "token required")
		return
	}
	id, _ := strconv.Atoi(r.PathValue("id"))
	a.mu.Lock()
	defer a.mu.Unlock()
	i, ok := a.find(id)
	if !ok {
		if a.has(UpdateMissingIs200) {
			writeJSON(w, http.StatusOK, map[string]any{})
			return
		}
		fail(w, http.StatusNotFound, "no such pet")
		return
	}
	var in petInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		fail(w, http.StatusBadRequest, "bad json")
		return
	}
	if !in.valid() && !a.has(NoUpdateValidation) {
		fail(w, http.StatusUnprocessableEntity, "invalid pet")
		return
	}
	if !a.has(UpdateIgnoresBody) {
		a.pets[i] = seed.Pet{ID: id, Name: in.Name, PriceCents: in.PriceCents, Status: in.Status, CategoryID: in.CategoryID}
	}
	writeJSON(w, http.StatusOK, pet(a.pets[i]))
}

func (a *App) createOrder(w http.ResponseWriter, r *http.Request) {
	if !a.authed(r, OpenOrderCreate) {
		fail(w, http.StatusUnauthorized, "token required")
		return
	}
	var in struct {
		PetID      int `json:"pet_id"`
		CustomerID int `json:"customer_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		fail(w, http.StatusBadRequest, "bad json")
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, ok := a.find(in.PetID); !ok && !a.has(OrderUnknownPetAccepted) {
		fail(w, http.StatusUnprocessableEntity, "unknown pet")
		return
	}
	if (in.CustomerID < 1 || in.CustomerID > seed.CustomerCount) && !a.has(OrderUnknownCustomerAccepted) {
		fail(w, http.StatusUnprocessableEntity, "unknown customer")
		return
	}
	o := order{ID: len(a.orders) + 1, PetID: in.PetID, CustomerID: in.CustomerID}
	a.orders = append(a.orders, o)
	writeJSON(w, http.StatusCreated, o)
}

func (a *App) getOrder(w http.ResponseWriter, r *http.Request) {
	if !a.authed(r, OpenOrderGet) {
		fail(w, http.StatusUnauthorized, "token required")
		return
	}
	id, _ := strconv.Atoi(r.PathValue("id"))
	a.mu.Lock()
	defer a.mu.Unlock()
	if id < 1 || id > len(a.orders) {
		fail(w, http.StatusNotFound, "no such order")
		return
	}
	writeJSON(w, http.StatusOK, a.orders[id-1])
}

const csrf = "csrf-token-1"

func (a *App) adminLoginPage(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintf(w, `<form method="post"><input type="hidden" name="csrf" value="%s"><input name="username"><input type="password" name="password"></form>`, csrf)
}

func (a *App) adminLogin(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	if r.Form.Get("csrf") != csrf || r.Form.Get("username") != Username || r.Form.Get("password") != Password {
		http.Error(w, "denied", http.StatusForbidden)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "admin", Value: "ok", Path: "/"})
	http.Redirect(w, r, "/admin/petshop/pet/", http.StatusFound)
}

func (a *App) adminList(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("admin")
	if (err != nil || c.Value != "ok") && !a.has(OpenAdmin) {
		http.Redirect(w, r, "/admin/login/", http.StatusFound)
		return
	}
	var b strings.Builder
	b.WriteString("<table>")
	if !a.has(AdminListEmpty) {
		a.mu.Lock()
		for _, p := range a.pets[:20] {
			fmt.Fprintf(&b, "<tr><td>%s</td></tr>", html.EscapeString(p.Name))
		}
		a.mu.Unlock()
	}
	b.WriteString("</table>")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(b.String()))
}
