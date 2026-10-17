package rest

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"github.com/angvp/tango-benchmarks/apps/togo/internal/app"
)

type shop struct {
	a      *app.App
	tokens *tokens
}

// RegisterShopRoutes mounts the pet shop's API on togo's Huma instance.
func RegisterShopRoutes(api huma.API, a *app.App) {
	s := &shop{a: a, tokens: newTokens()}
	auth := huma.Middlewares{s.requireToken(api)}

	huma.Register(api, huma.Operation{OperationID: "login", Method: http.MethodPost, Path: "/api/login",
		Summary: "Sign in as staff", Tags: []string{"auth"}}, s.login)

	huma.Register(api, huma.Operation{OperationID: "list-pets", Method: http.MethodGet, Path: "/api/pets",
		Summary: "List pets", Tags: []string{"pets"}}, s.listPets)
	huma.Register(api, huma.Operation{OperationID: "get-pet", Method: http.MethodGet, Path: "/api/pets/{id}",
		Summary: "Get a pet", Tags: []string{"pets"}}, s.getPet)
	huma.Register(api, huma.Operation{OperationID: "create-pet", Method: http.MethodPost, Path: "/api/pets",
		DefaultStatus: http.StatusCreated, Summary: "Create a pet", Tags: []string{"pets"}, Middlewares: auth}, s.createPet)
	huma.Register(api, huma.Operation{OperationID: "update-pet", Method: http.MethodPut, Path: "/api/pets/{id}",
		Summary: "Replace a pet", Tags: []string{"pets"}, Middlewares: auth}, s.updatePet)

	huma.Register(api, huma.Operation{OperationID: "create-order", Method: http.MethodPost, Path: "/api/orders",
		DefaultStatus: http.StatusCreated, Summary: "Create an order", Tags: []string{"orders"}, Middlewares: auth}, s.createOrder)
	huma.Register(api, huma.Operation{OperationID: "get-order", Method: http.MethodGet, Path: "/api/orders/{id}",
		Summary: "Get an order", Tags: []string{"orders"}, Middlewares: auth}, s.getOrder)
}
