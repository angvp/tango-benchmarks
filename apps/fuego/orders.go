package main

import (
	"errors"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/jackc/pgx/v5"

	"github.com/angvp/tango-benchmarks/apps/fuego/store"
)

type orderOutput struct {
	ID         int32 `json:"id"`
	PetID      int32 `json:"pet_id"`
	CustomerID int32 `json:"customer_id"`
}

type orderInput struct {
	PetID      int64 `json:"pet_id"`
	CustomerID int64 `json:"customer_id"`
}

func (sh *shop) createOrder(c fuego.ContextWithBody[orderInput]) (orderOutput, error) {
	in, err := c.Body()
	if err != nil {
		return orderOutput{}, err
	}
	unprocessable := failure(http.StatusUnprocessableEntity, "pet_id and customer_id must name existing rows")
	if in.PetID < 1 || in.PetID > maxInt32 || in.CustomerID < 1 || in.CustomerID > maxInt32 {
		return orderOutput{}, unprocessable
	}
	petOK, err := sh.q.PetExists(c.Context(), int32(in.PetID))
	if err != nil {
		return orderOutput{}, err
	}
	customerOK, err := sh.q.CustomerExists(c.Context(), int32(in.CustomerID))
	if err != nil {
		return orderOutput{}, err
	}
	if !petOK || !customerOK {
		return orderOutput{}, unprocessable
	}
	o, err := sh.q.CreateOrder(c.Context(), store.CreateOrderParams{PetID: int32(in.PetID), CustomerID: int32(in.CustomerID)})
	if err != nil {
		return orderOutput{}, err
	}
	return orderOutput{o.ID, o.PetID, o.CustomerID}, nil
}

func (sh *shop) getOrder(c fuego.ContextNoBody) (orderOutput, error) {
	id, ok := idParam(c)
	if !ok {
		return orderOutput{}, failure(http.StatusNotFound, "no such order")
	}
	o, err := sh.q.GetOrder(c.Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		return orderOutput{}, failure(http.StatusNotFound, "no such order")
	}
	if err != nil {
		return orderOutput{}, err
	}
	return orderOutput{o.ID, o.PetID, o.CustomerID}, nil
}
