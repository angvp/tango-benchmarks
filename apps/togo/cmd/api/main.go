// Command api is the shop HTTP entrypoint. It boots the shared togo stack
// (Huma REST + OpenAPI and gqlgen GraphQL on the kernel) and serves it.
package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/angvp/tango-benchmarks/apps/togo/internal/rest"
	"github.com/angvp/tango-benchmarks/apps/togo/internal/server"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "openapi" {
		b, err := server.OpenAPI()
		if err != nil {
			panic(err)
		}
		os.Stdout.Write(b)
		return
	}

	a := server.Boot()
	defer a.Kernel.Close()
	password := os.Getenv("BENCH_STAFF_PASSWORD")
	if password == "" {
		password = "bench-staff-password"
	}
	if n, err := strconv.Atoi(os.Getenv("DB_MAX_IDLE")); err == nil && n > 0 {
		a.SQLDB.SetMaxIdleConns(n) // the kernel keeps database/sql's default of 2 idle connections
	}
	if err := rest.EnsureStaff(context.Background(), a, password); err != nil {
		panic(err)
	}
	k := a.Kernel
	fmt.Printf("→ shop listening on %s  (GraphQL %s · REST %s · docs %s)\n",
		k.Config.Addr, k.Config.GraphQLPath, k.Config.RESTPath, k.Config.DocsPath)
	if err := k.Serve(context.Background()); err != nil {
		panic(err)
	}
}
