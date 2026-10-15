package main

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/angvp/tango"
	"github.com/angvp/tango/admin"
	"github.com/angvp/tango/db"

	"petshop/migrations"
	"petshop/project"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	if err := tango.LoadEnvFile(".env"); err != nil {
		return err
	}

	dsn, err := tango.LoadDBConfigFromEnv()
	if err != nil {
		return err
	}

	sqlDB, err := sql.Open(dsn.Driver, dsn.Source)
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	store := db.NewStore(sqlDB, dsn.Dialect)
	// The installed apps, middleware and other configuration live in
	// project/project.go, shared with the shell.
	config := project.Config(store)

	if handled, err := admin.HandleCLI(context.Background(), store, os.Args[1:], os.Stdin, os.Stdout, os.Stderr); handled || err != nil {
		return err
	}

	handled, err := tango.DispatchFlags(config, sqlDB, dsn.Dialect, migrations.Migrations)
	if handled || err != nil {
		return err
	}

	// Ctrl-C and SIGTERM (what hosting platforms send on a deploy) stop the
	// server gracefully: in-flight requests finish before the process exits.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	fmt.Println("listening on", config.Addr)
	if url := adminURL(config.Addr); url != "" {
		fmt.Println("admin:", url)
	}
	return tango.ServeContext(ctx, config, sqlDB, dsn.Dialect)
}

// adminURL is where the admin can be opened from this machine for the listen
// address addr, or "" when addr is not a host:port.
func adminURL(addr string) string {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return ""
	}
	switch host {
	case "", "0.0.0.0", "::":
		host = "localhost"
	}
	return "http://" + net.JoinHostPort(host, port) + "/admin/"
}
