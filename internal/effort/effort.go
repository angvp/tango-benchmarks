// Package effort measures how much it took to build an app: the steps, files,
// lines and tokens of what a developer wrote, kept apart from what the
// framework generated for them. docs/effort.md states the rules this code
// applies identically to every app.
package effort

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

type Category string

const (
	// Hand is a file the developer wrote.
	Hand Category = "hand_written"
	// ScaffoldEdited is a file a framework command created and the developer
	// then edited by hand; it is counted whole, and shown apart.
	ScaffoldEdited Category = "scaffold_edited"
	// Generated is a file a framework command wrote and nobody edited.
	Generated Category = "generated"
)

// Rules are an app's effort.json: which files are generated, and which
// scaffold files were edited. Everything else that counts is hand-written.
type Rules struct {
	Generated      []string `json:"generated"`
	ScaffoldEdited []string `json:"scaffold_edited"`
}

// ParseRules reads an effort.json, refusing unknown keys and invalid globs so
// a typo cannot quietly move a file into the wrong column.
func ParseRules(raw []byte) (Rules, error) {
	var r Rules
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&r); err != nil {
		return Rules{}, fmt.Errorf("effort.json: %w", err)
	}
	for _, pattern := range slices.Concat(r.Generated, r.ScaffoldEdited) {
		if !doublestar.ValidatePattern(pattern) {
			return Rules{}, fmt.Errorf("effort.json: %q is not a valid glob", pattern)
		}
	}
	return r, nil
}

// Files that package, run or describe the app for the benchmark rather than
// being part of the pet shop: they are the same kind of thing for every app,
// so none is counted. Dependency manifests are counted as dependencies.
var infrastructure = []string{
	"Dockerfile", "docker-entrypoint.sh", "compose.yaml", "app.json", "seed-map.json",
	"effort.json", "steps.txt", "CONFIGURATION.md", ".gitignore",
	"go.mod", "go.sum", "requirements.txt", "Gemfile", "Gemfile.lock", "composer.json", "composer.lock",
}

var noiseDirs = []string{"__pycache__", "bin", "node_modules", ".git", "vendor"}

// Classify says which column a file belongs in, and whether it counts at all.
// rel is the path relative to the app, with forward slashes.
func Classify(rel string, rules Rules) (Category, bool) {
	if slices.Contains(infrastructure, path.Base(rel)) && !strings.Contains(rel, "/") {
		return "", false
	}
	for _, segment := range strings.Split(path.Dir(rel), "/") {
		if slices.Contains(noiseDirs, segment) {
			return "", false
		}
	}
	if strings.HasSuffix(rel, ".pyc") {
		return "", false
	}
	for _, p := range rules.Generated {
		if ok, _ := doublestar.Match(p, rel); ok {
			return Generated, true
		}
	}
	for _, p := range rules.ScaffoldEdited {
		if ok, _ := doublestar.Match(p, rel); ok {
			return ScaffoldEdited, true
		}
	}
	return Hand, true
}

// CountLines counts the lines that are not blank. Comments count: a comment is
// something the developer wrote, and removing them would need a parser per
// language.
func CountLines(content []byte) int {
	n := 0
	for _, line := range bytes.Split(content, []byte("\n")) {
		if len(bytes.TrimSpace(line)) > 0 {
			n++
		}
	}
	return n
}

type Tokenizer interface {
	Count(text string) int
	Name() string
}

type Counts struct {
	Files  int `json:"files"`
	Lines  int `json:"lines"`
	Tokens int `json:"tokens"`
}

type Report struct {
	App            string `json:"app"`
	Tokenizer      string `json:"tokenizer"`
	Steps          int    `json:"steps"`
	Hand           Counts `json:"hand_written"`
	ScaffoldEdited Counts `json:"scaffold_edited"`
	Generated      Counts `json:"generated"`
	// Authored is hand-written plus edited scaffold: what the developer wrote
	// or edited. It sits beside the columns above and is not a score.
	Authored     Counts `json:"authored"`
	Dependencies int    `json:"dependencies"`
	ImageBytes   int64  `json:"image_bytes,omitempty"`
}

// Measure walks an app directory and counts files, lines and tokens per
// category. Binary files are skipped.
func Measure(dir string, rules Rules, tok Tokenizer) (Report, error) {
	var r Report
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if slices.Contains(noiseDirs, d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(dir, p)
		if err != nil {
			return err
		}
		category, counts := Classify(filepath.ToSlash(rel), rules)
		if !counts {
			return nil
		}
		content, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		if bytes.IndexByte(content, 0) >= 0 {
			return nil
		}
		target := &r.Hand
		switch category {
		case ScaffoldEdited:
			target = &r.ScaffoldEdited
		case Generated:
			target = &r.Generated
		}
		target.Files++
		target.Lines += CountLines(content)
		target.Tokens += tok.Count(string(content))
		return nil
	})
	r.Authored = Counts{
		Files:  r.Hand.Files + r.ScaffoldEdited.Files,
		Lines:  r.Hand.Lines + r.ScaffoldEdited.Lines,
		Tokens: r.Hand.Tokens + r.ScaffoldEdited.Tokens,
	}
	return r, err
}

// CountDependencies counts the direct dependencies an app declares, from its
// go.mod, requirements.txt, Gemfile or composer.json. An app with none of them
// is an error.
func CountDependencies(dir string) (int, error) {
	if raw, err := os.ReadFile(filepath.Join(dir, "go.mod")); err == nil {
		return countGoDependencies(string(raw)), nil
	}
	if raw, err := os.ReadFile(filepath.Join(dir, "requirements.txt")); err == nil {
		return countPythonDependencies(string(raw)), nil
	}
	if raw, err := os.ReadFile(filepath.Join(dir, "Gemfile")); err == nil {
		return countRubyDependencies(string(raw)), nil
	}
	if raw, err := os.ReadFile(filepath.Join(dir, "composer.json")); err == nil {
		return countPHPDependencies(raw)
	}
	return 0, fmt.Errorf("%s has no go.mod, requirements.txt, Gemfile or composer.json: add a rule for its ecosystem", dir)
}

func countGoDependencies(gomod string) int {
	n, inBlock := 0, false
	for _, line := range strings.Split(gomod, "\n") {
		line = strings.TrimSpace(line)
		switch {
		case line == "require (":
			inBlock = true
		case inBlock && line == ")":
			inBlock = false
		case inBlock && line != "" && !strings.Contains(line, "// indirect"):
			n++
		case strings.HasPrefix(line, "require ") && !strings.HasSuffix(line, "(") && !strings.Contains(line, "// indirect"):
			n++
		}
	}
	return n
}

// countRubyDependencies counts the gems a Gemfile requires for production:
// not comments, not gems limited to other platforms, not the development and
// test groups.
func countRubyDependencies(gemfile string) int {
	n, skipDepth := 0, 0
	for _, line := range strings.Split(gemfile, "\n") {
		line = strings.TrimSpace(line)
		switch {
		case skipDepth > 0 && line == "end":
			skipDepth--
		case skipDepth > 0 && strings.HasSuffix(line, " do"):
			skipDepth++
		case strings.HasPrefix(line, "group ") && (strings.Contains(line, ":development") || strings.Contains(line, ":test")):
			skipDepth = 1
		case skipDepth == 0 && strings.HasPrefix(line, "gem ") && !strings.Contains(line, "platforms:"):
			n++
		}
	}
	return n
}

// countPHPDependencies counts composer.json's require entries other than php
// and its extensions; require-dev is not counted.
func countPHPDependencies(raw []byte) (int, error) {
	var manifest struct {
		Require map[string]string `json:"require"`
	}
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return 0, fmt.Errorf("composer.json: %w", err)
	}
	n := 0
	for name := range manifest.Require {
		if name != "php" && !strings.HasPrefix(name, "ext-") {
			n++
		}
	}
	return n, nil
}

func countPythonDependencies(requirements string) int {
	n := 0
	for _, line := range strings.Split(requirements, "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "-") {
			n++
		}
	}
	return n
}

// CountSteps counts the commands in a steps file: one per line, blank lines
// and # comments ignored.
func CountSteps(file string) (int, error) {
	raw, err := os.ReadFile(file)
	if err != nil {
		return 0, fmt.Errorf("steps file: %w", err)
	}
	n := 0
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			n++
		}
	}
	return n, nil
}

// Run produces the whole report for one app directory. imageBytes is 0 when
// the image was not measured.
func Run(dir, name string, rules Rules, tok Tokenizer, imageBytes int64) (Report, error) {
	r, err := Measure(dir, rules, tok)
	if err != nil {
		return Report{}, err
	}
	if r.Steps, err = CountSteps(filepath.Join(dir, "steps.txt")); err != nil {
		return Report{}, err
	}
	if r.Dependencies, err = CountDependencies(dir); err != nil {
		return Report{}, err
	}
	r.App, r.Tokenizer, r.ImageBytes = name, tok.Name(), imageBytes
	return r, nil
}
