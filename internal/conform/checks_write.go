package conform

import (
	"fmt"
	"net/http"
)

var validPet = map[string]any{"name": "Conformance Pet", "price_cents": 12345, "status": "pending", "category_id": 2}

var authChecks = []check{
	{"login.bad", func(e *env) error {
		_, err := e.expect(http.StatusUnauthorized, "POST", "/api/login", false, map[string]string{"username": e.cfg.Username, "password": "definitely-wrong"})
		return err
	}},
	{"login.good", func(e *env) error {
		r, err := e.expect(http.StatusOK, "POST", "/api/login", false, map[string]string{"username": e.cfg.Username, "password": e.cfg.Password})
		if err != nil {
			return err
		}
		var out struct {
			Token string `json:"token"`
		}
		if err := decode(r, &out); err != nil {
			return err
		}
		if out.Token == "" {
			return fmt.Errorf("login returned no token")
		}
		e.token = out.Token
		return nil
	}},
	{"auth.pet_create_unauth", func(e *env) error {
		_, err := e.expect(http.StatusUnauthorized, "POST", "/api/pets", false, validPet)
		return err
	}},
	{"auth.pet_update_unauth", func(e *env) error {
		_, err := e.expect(http.StatusUnauthorized, "PUT", "/api/pets/1", false, validPet)
		return err
	}},
	{"auth.order_create_unauth", func(e *env) error {
		_, err := e.expect(http.StatusUnauthorized, "POST", "/api/orders", false, map[string]int{"pet_id": 1, "customer_id": 1})
		return err
	}},
	{"auth.order_get_unauth", func(e *env) error {
		_, err := e.expect(http.StatusUnauthorized, "GET", "/api/orders/1", false, nil)
		return err
	}},
}

func requireToken(e *env) error {
	if e.token == "" {
		return fmt.Errorf("no token: login.good did not pass")
	}
	return nil
}

var writeChecks = []check{
	{"pet.create", func(e *env) error {
		if err := requireToken(e); err != nil {
			return err
		}
		r, err := e.expect(http.StatusCreated, "POST", "/api/pets", true, validPet)
		if err != nil {
			return err
		}
		var p pet
		if err := decode(r, &p); err != nil {
			return err
		}
		if p.ID <= 10000 || p.Name != "Conformance Pet" || p.PriceCents != 12345 || p.Status != "pending" || p.CategoryID != 2 {
			return fmt.Errorf("created pet = %+v", p)
		}
		e.petID = p.ID
		return sameOnRead(e, p)
	}},
	{"pet.create_invalid", func(e *env) error {
		if err := requireToken(e); err != nil {
			return err
		}
		for name, body := range invalidPets() {
			if _, err := e.expect(http.StatusUnprocessableEntity, "POST", "/api/pets", true, body); err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
		}
		return nil
	}},
	{"pet.update", func(e *env) error {
		if e.petID == 0 {
			return fmt.Errorf("no pet to update: pet.create did not pass")
		}
		body := map[string]any{"name": "Renamed Pet", "price_cents": 999, "status": "sold", "category_id": 5}
		r, err := e.expect(http.StatusOK, "PUT", fmt.Sprintf("/api/pets/%d", e.petID), true, body)
		if err != nil {
			return err
		}
		var p pet
		if err := decode(r, &p); err != nil {
			return err
		}
		want := pet{e.petID, "Renamed Pet", 999, "sold", 5}
		if p != want {
			return fmt.Errorf("updated pet = %+v, want %+v", p, want)
		}
		return sameOnRead(e, want)
	}},
	{"pet.update_missing", func(e *env) error {
		if err := requireToken(e); err != nil {
			return err
		}
		_, err := e.expect(http.StatusNotFound, "PUT", "/api/pets/999999", true, validPet)
		return err
	}},
	{"pet.update_invalid", func(e *env) error {
		if e.petID == 0 {
			return fmt.Errorf("no pet to update: pet.create did not pass")
		}
		for name, body := range invalidPets() {
			if _, err := e.expect(http.StatusUnprocessableEntity, "PUT", fmt.Sprintf("/api/pets/%d", e.petID), true, body); err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
		}
		return nil
	}},
	{"order.create", func(e *env) error {
		if e.petID == 0 {
			return fmt.Errorf("no pet to order: pet.create did not pass")
		}
		r, err := e.expect(http.StatusCreated, "POST", "/api/orders", true, map[string]int{"pet_id": e.petID, "customer_id": 7})
		if err != nil {
			return err
		}
		var o order
		if err := decode(r, &o); err != nil {
			return err
		}
		if o.ID < 1 || o.PetID != e.petID || o.CustomerID != 7 {
			return fmt.Errorf("created order = %+v", o)
		}
		e.orderID = o.ID
		return nil
	}},
	{"order.create_unknown_pet", func(e *env) error {
		if err := requireToken(e); err != nil {
			return err
		}
		_, err := e.expect(http.StatusUnprocessableEntity, "POST", "/api/orders", true, map[string]int{"pet_id": 999999, "customer_id": 1})
		return err
	}},
	{"order.create_unknown_customer", func(e *env) error {
		if err := requireToken(e); err != nil {
			return err
		}
		_, err := e.expect(http.StatusUnprocessableEntity, "POST", "/api/orders", true, map[string]int{"pet_id": 1, "customer_id": 999999})
		return err
	}},
	{"order.get", func(e *env) error {
		if e.orderID == 0 {
			return fmt.Errorf("no order to read: order.create did not pass")
		}
		r, err := e.expect(http.StatusOK, "GET", fmt.Sprintf("/api/orders/%d", e.orderID), true, nil)
		if err != nil {
			return err
		}
		var o order
		if err := decode(r, &o); err != nil {
			return err
		}
		if o != (order{e.orderID, e.petID, 7}) {
			return fmt.Errorf("order read back = %+v", o)
		}
		return nil
	}},
	{"order.get_missing", func(e *env) error {
		if err := requireToken(e); err != nil {
			return err
		}
		_, err := e.expect(http.StatusNotFound, "GET", "/api/orders/999999", true, nil)
		return err
	}},
}

type order struct {
	ID         int `json:"id"`
	PetID      int `json:"pet_id"`
	CustomerID int `json:"customer_id"`
}

func sameOnRead(e *env, want pet) error {
	r, err := e.expect(http.StatusOK, "GET", fmt.Sprintf("/api/pets/%d", want.ID), false, nil)
	if err != nil {
		return err
	}
	var got pet
	if err := decode(r, &got); err != nil {
		return err
	}
	if got != want {
		return fmt.Errorf("read back %+v, want %+v", got, want)
	}
	return nil
}

func invalidPets() map[string]map[string]any {
	with := func(key string, value any) map[string]any {
		body := map[string]any{}
		for k, v := range validPet {
			body[k] = v
		}
		body[key] = value
		return body
	}
	return map[string]map[string]any{
		"empty name":       with("name", ""),
		"negative price":   with("price_cents", -1),
		"unknown status":   with("status", "adopted"),
		"unknown category": with("category_id", 999),
	}
}
