package shop

import (
	"net/http"

	"github.com/angvp/tango"
	"github.com/angvp/tango/admin"
	tangojwt "github.com/angvp/tango/auth/jwt"
	"github.com/angvp/tango/db"
)

// New is the pet shop: four models, the admin over them, and the JSON API.
func New(store *db.Store, tokens *tangojwt.Service) tango.App {
	return tango.NewApp("shop", func(registry *tango.Registry) error {
		for _, m := range []any{Category{}, Pet{}, Customer{}, Order{}} {
			if err := registry.Models().Register(m); err != nil {
				return err
			}
		}
		if err := registry.Admin().Register(Category{}, admin.Options{Label: "Name", ListDisplay: []string{"Name"}}); err != nil {
			return err
		}
		if err := registry.Admin().Register(Pet{}, admin.Options{
			ListDisplay: []string{"Name", "PriceCents", "Status", "CategoryID"},
			Search:      []string{"Name"},
			Ordering:    []string{"ID"},
			Filters:     []string{"CategoryID"},
		}); err != nil {
			return err
		}
		if err := registry.Admin().Register(Customer{}, admin.Options{Label: "Name", ListDisplay: []string{"Name", "Email"}, Search: []string{"Name", "Email"}}); err != nil {
			return err
		}
		if err := registry.Admin().Register(Order{}, admin.Options{ListDisplay: []string{"PetID", "CustomerID"}}); err != nil {
			return err
		}

		api, err := newAPI(registry, store, tokens)
		if err != nil {
			return err
		}
		auth := tokens.Middleware(tangojwt.BearerToken)
		return registry.Routes().Include("/api/", tango.URLs{
			tango.Path(http.MethodPost, "login", api.login),
			tango.Path(http.MethodGet, "pets", api.listPets),
			tango.Path(http.MethodGet, "pets/{id}", api.getPet),
			tango.Path(http.MethodPost, "pets", tangojwt.Require(api.createPet), tango.Use(auth)),
			tango.Path(http.MethodPut, "pets/{id}", tangojwt.Require(api.updatePet), tango.Use(auth)),
			tango.Path(http.MethodPost, "orders", tangojwt.Require(api.createOrder), tango.Use(auth)),
			tango.Path(http.MethodGet, "orders/{id}", tangojwt.Require(api.getOrder), tango.Use(auth)),
		})
	})
}
