package rest

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
)

type orderOutput struct {
	ID         int32 `json:"id"`
	PetID      int32 `json:"pet_id"`
	CustomerID int32 `json:"customer_id"`
}

type orderInput struct {
	Body struct {
		PetID      int64 `json:"pet_id"`
		CustomerID int64 `json:"customer_id"`
	}
}

type orderResponse struct{ Body orderOutput }

func (s *shop) exists(ctx context.Context, table string, id int64) (bool, error) {
	var ok bool
	// table is one of two constants below, never request data.
	err := s.a.SQLDB.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM `+table+` WHERE id = $1)`, id).Scan(&ok)
	return ok, err
}

func (s *shop) createOrder(ctx context.Context, in *orderInput) (*orderResponse, error) {
	unprocessable := failure(http.StatusUnprocessableEntity, "pet_id and customer_id must name existing rows")
	b := in.Body
	if b.PetID < 1 || b.PetID > maxInt32 || b.CustomerID < 1 || b.CustomerID > maxInt32 {
		return nil, unprocessable
	}
	petOK, err := s.exists(ctx, "pets", b.PetID)
	if err != nil {
		return nil, err
	}
	customerOK, err := s.exists(ctx, "customers", b.CustomerID)
	if err != nil {
		return nil, err
	}
	if !petOK || !customerOK {
		return nil, unprocessable
	}
	var o orderOutput
	err = s.a.SQLDB.QueryRowContext(ctx, `INSERT INTO orders (pet_id, customer_id) VALUES ($1, $2) RETURNING id, pet_id, customer_id`,
		b.PetID, b.CustomerID).Scan(&o.ID, &o.PetID, &o.CustomerID)
	return &orderResponse{Body: o}, err
}

func (s *shop) getOrder(ctx context.Context, in *idInput) (*orderResponse, error) {
	id, ok := idOf(in.ID)
	if !ok {
		return nil, failure(http.StatusNotFound, "no such order")
	}
	var o orderOutput
	err := s.a.SQLDB.QueryRowContext(ctx, `SELECT id, pet_id, customer_id FROM orders WHERE id = $1`, id).Scan(&o.ID, &o.PetID, &o.CustomerID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, failure(http.StatusNotFound, "no such order")
	}
	return &orderResponse{Body: o}, err
}
