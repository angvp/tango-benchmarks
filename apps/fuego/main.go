// The Fuego pet shop: net/http routing, typed handlers and the generated
// OpenAPI document come from Fuego; the data layer is the same sqlc store as
// the plain Go baseline, since Fuego has no ORM.
package main

import (
	"context"
	_ "embed"
	"errors"
	"log/slog"
	"os"

	"github.com/go-fuego/fuego"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/angvp/tango-benchmarks/apps/fuego/store"
)

//go:embed db/schema.sql
var schema string

const staffUsername = "bench-staff"

func main() {
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	if err := run(); err != nil {
		log.Error("stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return errors.New("DATABASE_URL is required")
	}
	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	password := os.Getenv("BENCH_STAFF_PASSWORD")
	if password == "" {
		password = "bench-staff-password"
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return err
	}
	defer pool.Close()
	if _, err := pool.Exec(ctx, schema); err != nil {
		return err
	}
	q := store.New(pool)
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := q.UpsertStaff(ctx, store.UpsertStaffParams{Username: staffUsername, PasswordHash: string(hash)}); err != nil {
		return err
	}

	s := fuego.NewServer(fuego.WithAddr(addr), fuego.WithErrorSerializer(sendError))
	newShop(q, newTokens()).routes(s)
	return s.Run()
}
