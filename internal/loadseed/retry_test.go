package loadseed

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestIsDeadlockRecognisesOnlyPostgreSQLDeadlocks(t *testing.T) {
	deadlock := fmt.Errorf("truncate: %w", &pgconn.PgError{Code: "40P01"})
	if !IsDeadlock(deadlock) {
		t.Error("a wrapped 40P01 is a deadlock")
	}
	if IsDeadlock(&pgconn.PgError{Code: "23505"}) {
		t.Error("a unique violation is not a deadlock")
	}
	if IsDeadlock(errors.New("connection refused")) {
		t.Error("a plain error is not a deadlock")
	}
}
