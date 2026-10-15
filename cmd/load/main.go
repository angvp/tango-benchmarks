// Command load runs the load scenarios against one app and summarises them.
//
//	load run -app apps/tango -scenario pet -run 1 -container tangobench-app-1 -out results/tango-load
//	load summarize -dir results/tango-load
//
// Every result is stamped "local, not publishable". The parameters come from
// load/params.json (see docs/load.md).
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/angvp/tango-benchmarks/internal/conform"
	"github.com/angvp/tango-benchmarks/internal/load"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: load run|summarize [flags]")
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "run":
		err = runCmd(os.Args[2:])
	case "summarize":
		err = summarizeCmd(os.Args[2:])
	default:
		err = fmt.Errorf("unknown command %q (want run or summarize)", os.Args[1])
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "load:", err)
		os.Exit(1)
	}
}

func runCmd(args []string) error {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	app := fs.String("app", "", "app directory, e.g. apps/tango (required)")
	scenario := fs.String("scenario", "", "list, pet, order, admin or mix (required)")
	runNo := fs.Int("run", 1, "run number, for the file name")
	params := fs.String("params", "load/params.json", "parameters file")
	baseURL := fs.String("base-url", "http://127.0.0.1:8080", "")
	container := fs.String("container", "", "app container: restarted for the cold start and sampled for memory")
	out := fs.String("out", "", "directory for the result file (required)")
	fs.Parse(args)
	if *app == "" || *scenario == "" || *out == "" {
		return fmt.Errorf("-app, -scenario and -out are required")
	}
	p, err := readParams(*params)
	if err != nil {
		return err
	}
	appCfg, err := readApp(*app)
	if err != nil {
		return err
	}
	rec := load.Record{App: appCfg.Name, Run: *runNo}
	if *container != "" {
		cold, err := load.ColdStart(*container, *baseURL+"/api/pets?page_size=1", 90*time.Second)
		if err != nil {
			return err
		}
		rec.ColdStartMs = float64(cold) / float64(time.Millisecond)
	}
	target := load.Target{BaseURL: *baseURL, Username: "bench-staff", Password: passwordFromEnv(), Admin: appCfg.Admin, AdminNAReason: appCfg.AdminNAReason}
	ctx, stop := context.WithCancel(context.Background())
	peak := make(chan float64, 1)
	go func() {
		if *container == "" {
			peak <- 0
			return
		}
		b, _ := load.PeakMemory(ctx, *container)
		peak <- b
	}()
	res, err := load.Run(context.Background(), p, target, *scenario)
	stop()
	rec.PeakMemMB = <-peak / (1 << 20)
	if err != nil {
		return err
	}
	rec.Result = res
	if err := os.MkdirAll(*out, 0o755); err != nil {
		return err
	}
	raw, _ := json.MarshalIndent(rec, "", "  ")
	file := filepath.Join(*out, fmt.Sprintf("%s-%d.json", *scenario, *runNo))
	fmt.Printf("%s run %d: %d ok, %d errors, %.0f req/s, p50 %.2f ms, p99 %.2f ms [%s]\n", *scenario, *runNo, res.OK, res.Errors, res.RPS, res.P50Ms, res.P99Ms, load.Stamp)
	return os.WriteFile(file, append(raw, '\n'), 0o644)
}

func summarizeCmd(args []string) error {
	fs := flag.NewFlagSet("summarize", flag.ExitOnError)
	dir := fs.String("dir", "", "directory of run files (required)")
	fs.Parse(args)
	if *dir == "" {
		return fmt.Errorf("-dir is required")
	}
	files, err := filepath.Glob(filepath.Join(*dir, "*-*.json"))
	if err != nil {
		return err
	}
	var recs []load.Record
	for _, f := range files {
		if strings.HasSuffix(f, "summary.json") {
			continue
		}
		raw, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		var r load.Record
		if err := json.Unmarshal(raw, &r); err != nil {
			return fmt.Errorf("%s: %w", f, err)
		}
		recs = append(recs, r)
	}
	sum, err := load.Summarize(recs)
	if err != nil {
		return err
	}
	fmt.Print(sum.Table())
	raw, _ := json.MarshalIndent(sum, "", "  ")
	return os.WriteFile(filepath.Join(*dir, "summary.json"), append(raw, '\n'), 0o644)
}

func readParams(path string) (load.Params, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return load.Params{}, err
	}
	return load.ParseParams(raw)
}

func readApp(dir string) (conform.App, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "app.json"))
	if err != nil {
		return conform.App{}, err
	}
	return conform.ParseAppJSON(raw)
}

func passwordFromEnv() string {
	if v := os.Getenv("BENCH_STAFF_PASSWORD"); v != "" {
		return v
	}
	return "bench-staff-password"
}
