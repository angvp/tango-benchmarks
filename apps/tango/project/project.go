package project

import (
	"os"

	"github.com/angvp/tango"
	"github.com/angvp/tango/admin"
	tangojwt "github.com/angvp/tango/auth/jwt"
	"github.com/angvp/tango/db"

	"petshop/apps/shop"
)

// Config composes the application: its installed apps, middleware and
// ordinary configuration. The server (main.go) and the shell (shell/main.go)
// both call it, so they always register the same apps and models. Add each
// new app to InstalledApps here.
//
// Config does not load the .env file or open a database: each process does
// that itself and passes the store in.
func Config(store *db.Store) tango.Config {
	tokens, err := newTokens()
	if err != nil {
		panic(err) // a malformed secret is a deployment mistake; fail at startup
	}
	// The address is TANGO_ADDR, else the PORT hosting platforms set, else :8000.
	config := tango.LoadConfigFromEnv(tango.WithPortFromEnv())
	config.InstalledApps = []tango.App{
		shop.New(store, tokens),
		admin.New(store), // after every app that registers models with the admin
	}
	config.Middleware = []tango.Middleware{
		tango.RequestID(),
		tango.Recoverer(),
		tango.AccessLogger(),
		// Request bodies are capped at 1 MiB. If this application later needs
		// large uploads, remove the global body-limit middleware and apply
		// `MaxBodySize` only to the route groups or routes that should remain
		// limited.
		tango.MaxBodySize(1 << 20),
	}
	// Global middleware also wraps requests no route matches, so 404s and
	// 405s are logged and counted too.
	config.MiddlewareScope = tango.MiddlewareScopeAll
	return config
}

// newTokens builds the JWT service for the API. BENCH_JWT_SECRET is a
// throwaway benchmark value, not a secret: the default keeps the harness
// self-contained.
func newTokens() (*tangojwt.Service, error) {
	secret := os.Getenv("BENCH_JWT_SECRET")
	if secret == "" {
		secret = "benchmark-only-jwt-secret-32-bytes-or-more"
	}
	return tangojwt.NewService(tangojwt.Key{ID: "bench-v1", Secret: []byte(secret)}, nil, "petshop", "petshop")
}
