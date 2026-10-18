package main

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"

	"github.com/angvp/tango-benchmarks/apps/gin/store"
)

type orderJSON struct {
	ID         int32 `json:"id"`
	PetID      int32 `json:"pet_id"`
	CustomerID int32 `json:"customer_id"`
}

func (s *server) createOrder(c *gin.Context) {
	var in struct {
		PetID      int64 `json:"pet_id"`
		CustomerID int64 `json:"customer_id"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		writeError(c, http.StatusBadRequest, "body is not valid JSON")
		return
	}
	if in.PetID < 1 || in.PetID > maxInt32 || in.CustomerID < 1 || in.CustomerID > maxInt32 {
		writeError(c, http.StatusUnprocessableEntity, "pet_id and customer_id must name existing rows")
		return
	}
	petOK, err := s.q.PetExists(c.Request.Context(), int32(in.PetID))
	if err != nil {
		writeError(c, http.StatusInternalServerError, "database error")
		return
	}
	customerOK, err := s.q.CustomerExists(c.Request.Context(), int32(in.CustomerID))
	if err != nil {
		writeError(c, http.StatusInternalServerError, "database error")
		return
	}
	if !petOK || !customerOK {
		writeError(c, http.StatusUnprocessableEntity, "pet_id and customer_id must name existing rows")
		return
	}
	o, err := s.q.CreateOrder(c.Request.Context(), store.CreateOrderParams{PetID: int32(in.PetID), CustomerID: int32(in.CustomerID)})
	if err != nil {
		writeError(c, http.StatusInternalServerError, "database error")
		return
	}
	c.JSON(http.StatusCreated, orderJSON{o.ID, o.PetID, o.CustomerID})
}

func (s *server) getOrder(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		writeError(c, http.StatusNotFound, "no such order")
		return
	}
	o, err := s.q.GetOrder(c.Request.Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(c, http.StatusNotFound, "no such order")
		return
	}
	if err != nil {
		writeError(c, http.StatusInternalServerError, "database error")
		return
	}
	c.JSON(http.StatusOK, orderJSON{o.ID, o.PetID, o.CustomerID})
}
