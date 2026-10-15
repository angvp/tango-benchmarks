package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/uadmin/uadmin"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

func registerAPI() {
	http.HandleFunc("POST /api/login", login)
	http.HandleFunc("GET /api/pets", listPets)
	http.HandleFunc("GET /api/pets/{id}", getPet)
	http.HandleFunc("POST /api/pets", authenticated(createPet))
	http.HandleFunc("PUT /api/pets/{id}", authenticated(updatePet))
	http.HandleFunc("POST /api/orders", authenticated(createOrder))
	http.HandleFunc("GET /api/orders/{id}", authenticated(getOrder))
}

type petJSON struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	PriceCents int    `json:"price_cents"`
	Status     string `json:"status"`
	CategoryID uint   `json:"category_id"`
}

func toPetJSON(p Pet) petJSON {
	return petJSON{p.ID, p.Name, p.PriceCents, p.Status, p.CategoryID}
}

type orderJSON struct {
	ID         uint `json:"id"`
	PetID      uint `json:"pet_id"`
	CustomerID uint `json:"customer_id"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func fail(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func validStatus(s string) bool { return s == "available" || s == "pending" || s == "sold" }

func login(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		fail(w, http.StatusBadRequest, "body is not valid JSON")
		return
	}
	session, otpRequired := uadmin.Login(r, strings.ToLower(in.Username), in.Password)
	if session == nil || otpRequired || !session.User.Admin {
		fail(w, http.StatusUnauthorized, "bad credentials")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"token": session.Key})
}

// authenticated accepts the session key that login returned as a bearer token.
func authenticated(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok || key == "" {
			fail(w, http.StatusUnauthorized, "a valid bearer token is required")
			return
		}
		r.AddCookie(&http.Cookie{Name: "session", Value: key})
		if uadmin.IsAuthenticated(r) == nil {
			fail(w, http.StatusUnauthorized, "a valid bearer token is required")
			return
		}
		next(w, r)
	}
}

// intParam reads an optional integer query parameter: (value, present, ok).
func intParam(r *http.Request, name string) (int, bool, bool) {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		return 0, false, true
	}
	n, err := strconv.Atoi(raw)
	return n, true, err == nil
}

func listPets(w http.ResponseWriter, r *http.Request) {
	category, hasCategory, ok1 := intParam(r, "category")
	minPrice, hasMin, ok2 := intParam(r, "min_price")
	maxPrice, hasMax, ok3 := intParam(r, "max_price")
	page, hasPage, ok4 := intParam(r, "page")
	size, hasSize, ok5 := intParam(r, "page_size")
	if !(ok1 && ok2 && ok3 && ok4 && ok5) {
		fail(w, http.StatusBadRequest, "numeric parameters must be integers")
		return
	}
	if !hasPage {
		page = 1
	}
	if !hasSize {
		size = defaultPageSize
	}
	if page < 1 || size < 1 || size > maxPageSize {
		fail(w, http.StatusBadRequest, "page must be at least 1 and page_size between 1 and 100")
		return
	}
	status := r.URL.Query().Get("status")
	if status != "" && !validStatus(status) {
		fail(w, http.StatusBadRequest, "unknown status")
		return
	}
	conditions, args := []string{"1 = 1"}, []any{}
	if hasCategory {
		conditions, args = append(conditions, "category_id = ?"), append(args, category)
	}
	if status != "" {
		conditions, args = append(conditions, "status = ?"), append(args, status)
	}
	if hasMin {
		conditions, args = append(conditions, "price_cents >= ?"), append(args, minPrice)
	}
	if hasMax {
		conditions, args = append(conditions, "price_cents <= ?"), append(args, maxPrice)
	}
	query := strings.Join(conditions, " AND ")
	total := uadmin.Count(&Pet{}, query, args...)
	var pets []Pet
	if err := uadmin.AdminPage("id", true, (page-1)*size, size, &pets, query, args...); err != nil {
		fail(w, http.StatusInternalServerError, "database error")
		return
	}
	items := make([]petJSON, 0, len(pets))
	for _, p := range pets {
		items = append(items, toPetJSON(p))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "page": page, "page_size": size, "total": total})
}

func idPath(r *http.Request) (uint, bool) {
	n, err := strconv.ParseUint(r.PathValue("id"), 10, 32)
	return uint(n), err == nil && n > 0
}

func getPet(w http.ResponseWriter, r *http.Request) {
	id, ok := idPath(r)
	var pet Pet
	if !ok || uadmin.Get(&pet, "id = ?", id) != nil {
		fail(w, http.StatusNotFound, "no such pet")
		return
	}
	writeJSON(w, http.StatusOK, toPetJSON(pet))
}

type petInput struct {
	Name       string `json:"name"`
	PriceCents int64  `json:"price_cents"`
	Status     string `json:"status"`
	CategoryID int64  `json:"category_id"`
}

// problem names what is wrong with the input, or "" when it is valid.
func (in petInput) problem() string {
	switch {
	case in.Name == "":
		return "name is required"
	case in.PriceCents < 0 || in.PriceCents > 2147483647:
		return "price_cents must be between 0 and 2147483647"
	case !validStatus(in.Status):
		return "status must be available, pending or sold"
	case in.CategoryID < 1 || uadmin.Count(&Category{}, "id = ?", in.CategoryID) == 0:
		return "category_id names no category"
	}
	return ""
}

func createPet(w http.ResponseWriter, r *http.Request) {
	var in petInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		fail(w, http.StatusBadRequest, "body is not valid JSON")
		return
	}
	if msg := in.problem(); msg != "" {
		fail(w, http.StatusUnprocessableEntity, msg)
		return
	}
	pet := Pet{Name: in.Name, PriceCents: int(in.PriceCents), Status: in.Status, CategoryID: uint(in.CategoryID)}
	if err := uadmin.Save(&pet); err != nil {
		fail(w, http.StatusInternalServerError, "database error")
		return
	}
	writeJSON(w, http.StatusCreated, toPetJSON(pet))
}

func updatePet(w http.ResponseWriter, r *http.Request) {
	id, ok := idPath(r)
	var pet Pet
	if !ok || uadmin.Get(&pet, "id = ?", id) != nil {
		fail(w, http.StatusNotFound, "no such pet")
		return
	}
	var in petInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		fail(w, http.StatusBadRequest, "body is not valid JSON")
		return
	}
	if msg := in.problem(); msg != "" {
		fail(w, http.StatusUnprocessableEntity, msg)
		return
	}
	pet = Pet{Model: pet.Model, Name: in.Name, PriceCents: int(in.PriceCents), Status: in.Status, CategoryID: uint(in.CategoryID)}
	if err := uadmin.Save(&pet); err != nil {
		fail(w, http.StatusInternalServerError, "database error")
		return
	}
	writeJSON(w, http.StatusOK, toPetJSON(pet))
}

func createOrder(w http.ResponseWriter, r *http.Request) {
	var in struct {
		PetID      int64 `json:"pet_id"`
		CustomerID int64 `json:"customer_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		fail(w, http.StatusBadRequest, "body is not valid JSON")
		return
	}
	if in.PetID < 1 || in.CustomerID < 1 || uadmin.Count(&Pet{}, "id = ?", in.PetID) == 0 || uadmin.Count(&Customer{}, "id = ?", in.CustomerID) == 0 {
		fail(w, http.StatusUnprocessableEntity, "pet_id and customer_id must name existing rows")
		return
	}
	order := Order{PetID: uint(in.PetID), CustomerID: uint(in.CustomerID)}
	if err := uadmin.Save(&order); err != nil {
		fail(w, http.StatusInternalServerError, "database error")
		return
	}
	writeJSON(w, http.StatusCreated, orderJSON{order.ID, order.PetID, order.CustomerID})
}

func getOrder(w http.ResponseWriter, r *http.Request) {
	id, ok := idPath(r)
	var order Order
	if !ok || uadmin.Get(&order, "id = ?", id) != nil {
		fail(w, http.StatusNotFound, "no such order")
		return
	}
	writeJSON(w, http.StatusOK, orderJSON{order.ID, order.PetID, order.CustomerID})
}
