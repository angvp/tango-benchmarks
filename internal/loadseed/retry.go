package loadseed

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// deadlockDetected is PostgreSQL's SQLSTATE for a transaction it chose to
// abort to break a deadlock; the documented response is to run it again.
const deadlockDetected = "40P01"

// IsDeadlock reports whether err is PostgreSQL aborting a transaction because
// it deadlocked with another. A reset can meet this when the app is still
// finishing a request from the run before.
func IsDeadlock(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == deadlockDetected
}
