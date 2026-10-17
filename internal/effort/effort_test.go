package effort

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// words counts whitespace-separated words: a tokenizer stand-in so the
// counting rules can be tested without depending on a vocabulary.
type words struct{}

func (words) Count(text string) int { return len(strings.Fields(text)) }
func (words) Name() string          { return "words" }

func write(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestClassify(t *testing.T) {
	rules := Rules{
		Generated:      []string{"store/**", "**/migrations/*.py"},
		ScaffoldEdited: []string{"petshop/settings.py"},
	}
	cases := []struct {
		path string
		want Category
		in   bool
	}{
		{"main.go", Hand, true},
		{"apps/shop/api.go", Hand, true},
		{"store/queries.sql.go", Generated, true},
		{"shop/migrations/0001_initial.py", Generated, true},
		{"petshop/settings.py", ScaffoldEdited, true},
		{"Dockerfile", "", false},
		{"compose.yaml", "", false},
		{"app.json", "", false},
		{"seed-map.json", "", false},
		{"effort.json", "", false},
		{"steps.txt", "", false},
		{"CONFIGURATION.md", "", false},
		{"go.sum", "", false},
		{".gitignore", "", false},
		{"docker-entrypoint.sh", "", false},
		{"shop/__pycache__/views.cpython-313.pyc", "", false},
		{"bin/petshop", "", false},
		{"vendor/laravel/framework/src/Foo.php", "", false},
	}
	for _, tc := range cases {
		got, in := Classify(tc.path, rules)
		if in != tc.in || got != tc.want {
			t.Errorf("Classify(%q) = %q, %v; want %q, %v", tc.path, got, in, tc.want, tc.in)
		}
	}
}

func TestCountLinesIgnoresBlankLines(t *testing.T) {
	if got := CountLines([]byte("a\n\n  \n\tb\nc")); got != 3 {
		t.Fatalf("got %d, want 3", got)
	}
	if got := CountLines(nil); got != 0 {
		t.Fatalf("empty file: got %d, want 0", got)
	}
}

func TestMeasureSeparatesTheThreeKinds(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "main.go", "package main\n\nfunc main() {}\n") // hand: 2 lines, 5 words
	write(t, dir, "apps/a.go", "one two\nthree\n")               // hand: 2 lines, 3 words
	write(t, dir, "store/gen.go", "gen gen gen\n")               // generated: 1 line, 3 words
	write(t, dir, "settings.py", "x = 1\ny = 2\n\nz = 3\n")      // scaffold edited: 3 lines, 9 words
	write(t, dir, "Dockerfile", "FROM scratch\n")                // infrastructure: not counted
	write(t, dir, "shop/__pycache__/v.pyc", "junk junk")         // noise: not counted
	rules := Rules{Generated: []string{"store/**"}, ScaffoldEdited: []string{"settings.py"}}

	r, err := Measure(dir, rules, words{})
	if err != nil {
		t.Fatal(err)
	}
	if r.Hand != (Counts{Files: 2, Lines: 4, Tokens: 8}) {
		t.Errorf("hand = %+v, want 2 files, 4 lines, 8 tokens", r.Hand)
	}
	if r.Generated != (Counts{Files: 1, Lines: 1, Tokens: 3}) {
		t.Errorf("generated = %+v, want 1 file, 1 line, 3 tokens", r.Generated)
	}
	if r.ScaffoldEdited != (Counts{Files: 1, Lines: 3, Tokens: 9}) {
		t.Errorf("scaffold edited = %+v, want 1 file, 3 lines, 9 tokens", r.ScaffoldEdited)
	}
	// Authored is what the developer wrote or edited: hand-written plus edited
	// scaffold, never the generated files.
	if r.Authored != (Counts{Files: 3, Lines: 7, Tokens: 17}) {
		t.Errorf("authored = %+v, want 3 files, 7 lines, 17 tokens", r.Authored)
	}
}

func TestMeasureSkipsBinaryFiles(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "logo.bin", "ab\x00cd\n")
	r, err := Measure(dir, Rules{}, words{})
	if err != nil {
		t.Fatal(err)
	}
	if r.Hand.Files != 0 {
		t.Fatalf("a binary file must not count, got %+v", r.Hand)
	}
}

func TestDirectDependenciesGo(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "go.mod", `module x

go 1.24

require github.com/a/a v1.0.0

require (
	github.com/b/b v1.0.0
	github.com/c/c v1.0.0 // indirect
	github.com/d/d v1.0.0
)
`)
	got, err := CountDependencies(dir)
	if err != nil || got != 3 {
		t.Fatalf("got %d, %v; want 3 direct (a, b, d)", got, err)
	}
}

func TestDirectDependenciesPython(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "requirements.txt", "# comment\nDjango==6.1.2\n\ngunicorn==26.2.0\npsycopg[binary]==3.3.6\n")
	got, err := CountDependencies(dir)
	if err != nil || got != 3 {
		t.Fatalf("got %d, %v; want 3", got, err)
	}
}

func TestNoDependencyFileIsAnError(t *testing.T) {
	if _, err := CountDependencies(t.TempDir()); err == nil {
		t.Fatal("an app with no go.mod or requirements.txt must be reported, not counted as zero")
	}
}

func TestStepsIgnoreBlankLinesAndComments(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "steps.txt", "# from an empty directory\ngo install x\n\ntango newproject petshop\n   \n# done\nrun\n")
	got, err := CountSteps(filepath.Join(dir, "steps.txt"))
	if err != nil || got != 3 {
		t.Fatalf("got %d, %v; want 3", got, err)
	}
}

func TestParseRulesRejectsUnknownKeys(t *testing.T) {
	if _, err := ParseRules([]byte(`{"generated":["a/**"],"scaffold_edit":["b"]}`)); err == nil {
		t.Fatal("a misspelt key must be refused, not ignored")
	}
	if _, err := ParseRules([]byte(`{"generated":["[unclosed"]}`)); err == nil {
		t.Fatal("an invalid glob must be refused")
	}
}

func TestCL100KCountsKnownStrings(t *testing.T) {
	tok, err := NewCL100K()
	if err != nil {
		t.Fatal(err)
	}
	// Worked by hand from the published vocabulary: "hello" and " world" are
	// one token each; "package", " main" and the newline are three.
	if got := tok.Count("hello world"); got != 2 {
		t.Errorf("hello world = %d tokens, want 2", got)
	}
	if got := tok.Count("package main\n"); got != 3 {
		t.Errorf("package main = %d tokens, want 3", got)
	}
	if tok.Name() != "cl100k_base" {
		t.Errorf("name = %q", tok.Name())
	}
}

func TestReportIsStampedAndStable(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "main.go", "package main\n")
	write(t, dir, "go.mod", "module x\n\ngo 1.24\n\nrequire github.com/a/a v1.0.0\n")
	write(t, dir, "steps.txt", "build\nrun\n")
	r, err := Run(dir, "demo", Rules{}, words{}, 0)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(r)
	for _, want := range []string{`"app":"demo"`, `"tokenizer":"words"`, `"steps":2`, `"dependencies":1`} {
		if !strings.Contains(string(raw), want) {
			t.Errorf("report %s lacks %s", raw, want)
		}
	}
	if strings.Contains(string(raw), "image_bytes") {
		t.Errorf("an unmeasured image size must be omitted, got %s", raw)
	}
}

func TestDirectDependenciesRuby(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "Gemfile", `source "https://rubygems.org"
# gem "commented", "~> 1.0"
gem "rails", "~> 8.1.4"
gem "pg", "~> 1.1"
gem "tzinfo-data", platforms: %i[ windows jruby ]

group :development, :test do
  gem "debug"
end
gem "puma"
`)
	got, err := CountDependencies(dir)
	if err != nil || got != 3 {
		t.Fatalf("got %d, %v; want 3 (rails, pg, puma: no comment, platform gem or development group)", got, err)
	}
}

func TestDirectDependenciesPHP(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "composer.json", `{
  "require": {"php": "^8.3", "ext-pdo": "*", "laravel/framework": "^13.0", "laravel/sanctum": "^4.0"},
  "require-dev": {"phpunit/phpunit": "^12.0"}
}`)
	got, err := CountDependencies(dir)
	if err != nil || got != 2 {
		t.Fatalf("got %d, %v; want 2 (the framework and sanctum: not php, extensions or require-dev)", got, err)
	}
}

func TestPackageManagerFilesAreInfrastructure(t *testing.T) {
	for _, name := range []string{"Gemfile", "Gemfile.lock", "composer.json", "composer.lock"} {
		if _, in := Classify(name, Rules{}); in {
			t.Errorf("%s must not count as authored code", name)
		}
	}
}
