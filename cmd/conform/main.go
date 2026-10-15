// Command conform runs the pet shop conformance suite against a running app:
//
//	conform -base-url http://localhost:8080 -app apps/go-baseline/app.json -out results/go-baseline.json
//
// It exits 1 if any check fails. N/A is not a failure.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/angvp/tango-benchmarks/internal/conform"
)

func main() {
	baseURL := flag.String("base-url", "http://localhost:8080", "where the app listens")
	appPath := flag.String("app", "", "the app's app.json (required)")
	out := flag.String("out", "", "write the results as JSON to this file")
	flag.Parse()
	if *appPath == "" {
		fmt.Fprintln(os.Stderr, "conform: -app is required")
		os.Exit(2)
	}
	raw, err := os.ReadFile(*appPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "conform:", err)
		os.Exit(2)
	}
	app, err := conform.ParseAppJSON(raw)
	if err != nil {
		fmt.Fprintln(os.Stderr, "conform:", err)
		os.Exit(2)
	}
	password := os.Getenv("BENCH_STAFF_PASSWORD")
	if password == "" {
		password = "bench-staff-password"
	}
	results := conform.Run(context.Background(), conform.Config{
		BaseURL: *baseURL, Username: "bench-staff", Password: password,
		Admin: app.Admin, AdminNAReason: app.AdminNAReason,
	})
	pass, fail, na := 0, 0, 0
	for _, r := range results {
		switch r.Status {
		case conform.Pass:
			pass++
		case conform.Fail:
			fail++
			fmt.Printf("FAIL %s: %s\n", r.ID, r.Detail)
		case conform.NotApplicable:
			na++
			fmt.Printf("N/A  %s: %s\n", r.ID, r.Detail)
		}
	}
	fmt.Printf("%s: %d passed, %d failed, %d not applicable\n", app.Name, pass, fail, na)
	if *out != "" {
		if err := writeResults(*out, app.Name, results); err != nil {
			fmt.Fprintln(os.Stderr, "conform:", err)
			os.Exit(2)
		}
	}
	if fail > 0 {
		os.Exit(1)
	}
}

func writeResults(path, name string, results []conform.Result) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(map[string]any{"app": name, "checks": results}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(raw, '\n'), 0o644)
}
