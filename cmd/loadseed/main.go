// Command loadseed resets an app's pet shop tables and loads the shared seed
// into them:
//
//	loadseed -dsn "$DATABASE_URL" -map apps/go-baseline/seed-map.json
//
// The tables must already exist: each app creates its own schema. The DSN is
// never printed.
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v5"

	"github.com/angvp/tango-benchmarks/internal/loadseed"
)

func main() {
	dsn := flag.String("dsn", os.Getenv("DATABASE_URL"), "PostgreSQL URL (default: $DATABASE_URL)")
	mapPath := flag.String("map", "", "the app's seed-map.json (required)")
	dataDir := flag.String("data", "seed/data", "directory with the seed CSV files")
	flag.Parse()
	if *mapPath == "" || *dsn == "" {
		fmt.Fprintln(os.Stderr, "loadseed: -map and a DSN (-dsn or DATABASE_URL) are required")
		os.Exit(2)
	}
	if err := run(context.Background(), *dsn, *mapPath, *dataDir); err != nil {
		fmt.Fprintln(os.Stderr, "loadseed:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, dsn, mapPath, dataDir string) (err error) {
	raw, err := os.ReadFile(mapPath)
	if err != nil {
		return err
	}
	m, err := loadseed.ParseMap(raw)
	if err != nil {
		return err
	}
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return fmt.Errorf("connect: %w", err) // pgx errors do not echo the password
	}
	defer conn.Close(ctx)
	if _, err := conn.Exec(ctx, "BEGIN"); err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_, _ = conn.Exec(ctx, "ROLLBACK")
		}
	}()
	if _, err := conn.Exec(ctx, m.TruncateSQL()); err != nil {
		return fmt.Errorf("truncate: %w", err)
	}
	for _, t := range m.Tables {
		if err := load(ctx, conn, t, filepath.Join(dataDir, t.CSV)); err != nil {
			return err
		}
	}
	if _, err := conn.Exec(ctx, "COMMIT"); err != nil {
		return err
	}
	fmt.Printf("loaded %d tables\n", len(m.Tables))
	return nil
}

func load(ctx context.Context, conn *pgx.Conn, t loadseed.Table, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	header, err := bufio.NewReader(f).ReadString('\n')
	if err != nil {
		return fmt.Errorf("%s: reading header: %w", path, err)
	}
	stmt, err := t.CopySQL(header)
	if err != nil {
		return err
	}
	if _, err := f.Seek(0, 0); err != nil {
		return err
	}
	if _, err := conn.PgConn().CopyFrom(ctx, f, stmt); err != nil {
		return fmt.Errorf("copy %s: %w", t.Table, err)
	}
	if _, err := conn.Exec(ctx, t.SetvalSQL()); err != nil {
		return fmt.Errorf("reset sequence of %s: %w", t.Table, err)
	}
	return nil
}
