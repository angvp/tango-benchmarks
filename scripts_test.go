package tangobenchmarks_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func guard(t *testing.T, files map[string]string) error {
	t.Helper()
	script, err := filepath.Abs("scripts/check-no-results.sh")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	run := func(name string, args ...string) {
		cmd := exec.Command(name, args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%s %v: %v\n%s", name, args, err, out)
		}
	}
	run("git", "init", "-q")
	for path, content := range files {
		full := filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	run("git", "add", "-A")
	cmd := exec.Command("sh", script)
	cmd.Dir = dir
	return cmd.Run()
}

func TestGuardPassesACleanRepository(t *testing.T) {
	if err := guard(t, map[string]string{"README.md": "x", "apps/a/effort.json": "{}", "docs/load.md": "stamp: local, not publishable"}); err != nil {
		t.Errorf("clean repo rejected: %v", err)
	}
}

func TestGuardRefusesCommittedResults(t *testing.T) {
	for name, files := range map[string]map[string]string{
		"results dir":    {"results/tango.json": "{}"},
		"load summary":   {"x/summary.json": "{}"},
		"effort report":  {"y-effort.json": "{}"},
		"stamped result": {"out/run.json": `{"stamp": "local, not publishable"}`},
	} {
		if err := guard(t, files); err == nil {
			t.Errorf("%s: committed result was accepted", name)
		}
	}
}
