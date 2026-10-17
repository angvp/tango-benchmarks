// Command seed populates the database with sample data via the generated seeders.
// Run with `togo seed`.
package main

import (
	"context"
	"os"

	"github.com/togo-framework/togo"

	"github.com/angvp/tango-benchmarks/apps/togo/internal/app"
	"github.com/angvp/tango-benchmarks/apps/togo/internal/db/seeders"
	_ "github.com/angvp/tango-benchmarks/apps/togo/internal/plugins" // registers the db driver and other installed plugins
)

func main() {
	k := togo.New()
	ctx := context.Background()
	a := app.New(ctx, k)
	defer k.Close()

	// Exit non-zero when nothing was seeded, so scripts and CI see the failure.
	if a.SQLDB == nil {
		k.Log.Error("seed: database unavailable")
		os.Exit(1)
	}
	if err := seeders.SeedAll(ctx, a); err != nil {
		k.Log.Error("seed failed", "err", err)
		k.Close() // os.Exit skips the deferred Close
		os.Exit(1)
	}
	k.Log.Info("seed complete")
}
