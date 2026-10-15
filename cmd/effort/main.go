// Command effort reports how much it took to build one app: steps, files,
// lines and tokens of hand-written, edited-scaffold and generated code,
// dependencies and (when asked) image size. The rules are in docs/effort.md.
//
//	effort -app apps/tango [-image tangobench-app] [-out results/tango-effort.json]
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/angvp/tango-benchmarks/internal/effort"
)

func main() {
	app := flag.String("app", "", "the app directory, e.g. apps/tango (required)")
	image := flag.String("image", "", "a built Docker image whose size to report")
	out := flag.String("out", "", "write the report as JSON to this file")
	flag.Parse()
	if *app == "" {
		fmt.Fprintln(os.Stderr, "effort: -app is required")
		os.Exit(2)
	}
	if err := run(*app, *image, *out); err != nil {
		fmt.Fprintln(os.Stderr, "effort:", err)
		os.Exit(1)
	}
}

func run(app, image, out string) error {
	raw, err := os.ReadFile(filepath.Join(app, "effort.json"))
	if err != nil {
		return err
	}
	rules, err := effort.ParseRules(raw)
	if err != nil {
		return err
	}
	tok, err := effort.NewCL100K()
	if err != nil {
		return err
	}
	var size int64
	if image != "" {
		if size, err = imageBytes(image); err != nil {
			return err
		}
	}
	report, err := effort.Run(app, filepath.Base(app), rules, tok, size)
	if err != nil {
		return err
	}
	fmt.Printf("%s (%s)\n  steps         %d\n  authored      %d files, %d lines, %d tokens (hand-written %d files, edited scaffold %d files)\n  generated     %d files, %d lines, %d tokens\n  dependencies  %d\n",
		report.App, report.Tokenizer, report.Steps,
		report.Authored.Files, report.Authored.Lines, report.Authored.Tokens, report.Hand.Files, report.ScaffoldEdited.Files,
		report.Generated.Files, report.Generated.Lines, report.Generated.Tokens, report.Dependencies)
	if size > 0 {
		fmt.Printf("  image         %.1f MB\n", float64(size)/1e6)
	}
	if out == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return err
	}
	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(out, append(encoded, '\n'), 0o644)
}

func imageBytes(image string) (int64, error) {
	raw, err := exec.Command("docker", "image", "inspect", image, "--format", "{{.Size}}").Output()
	if err != nil {
		return 0, fmt.Errorf("docker image inspect %s: %w", image, err)
	}
	return strconv.ParseInt(strings.TrimSpace(string(raw)), 10, 64)
}
