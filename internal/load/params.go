package load

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"
)

// Stamp is written into every result until a run is declared publishable.
const Stamp = "local, not publishable"

// MinRuns is the fewest repetitions that make a spread meaningful.
const MinRuns = 5

// Scenarios lists the scenarios, in the order they run.
var Scenarios = []string{"list", "pet", "order", "admin", "mix"}

// SeedShape is the size of the seeded data, so requests name rows that exist.
type SeedShape struct {
	Categories int `json:"categories"`
	Pets       int `json:"pets"`
	Customers  int `json:"customers"`
}

// Params is the content of load/params.json: everything about a run that is
// not typed by hand on a command line.
type Params struct {
	WarmupSeconds   int       `json:"warmup_seconds"`
	DurationSeconds int       `json:"duration_seconds"`
	Concurrency     int       `json:"concurrency"`
	Runs            int       `json:"runs"`
	Seed            SeedShape `json:"seed"`
	ListPageSize    int       `json:"list_page_size"`
	ReadPercent     int       `json:"read_percent"`
	Scenarios       []string  `json:"scenarios"`
}

// ParseParams reads and validates params.json. Unknown keys are refused.
func ParseParams(raw []byte) (Params, error) {
	var p Params
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		return Params{}, fmt.Errorf("params.json: %w", err)
	}
	switch {
	case p.Concurrency < 1:
		return Params{}, fmt.Errorf("params.json: concurrency must be at least 1")
	case p.DurationSeconds < 1 || p.WarmupSeconds < 0:
		return Params{}, fmt.Errorf("params.json: duration_seconds must be at least 1 and warmup_seconds not negative")
	case p.Runs < MinRuns:
		return Params{}, fmt.Errorf("params.json: at least %d runs are needed for a spread, got %d", MinRuns, p.Runs)
	case p.ReadPercent < 0 || p.ReadPercent > 100:
		return Params{}, fmt.Errorf("params.json: read_percent must be 0-100")
	case p.ListPageSize < 1 || p.ListPageSize > 100:
		return Params{}, fmt.Errorf("params.json: list_page_size must be 1-100")
	case p.Seed.Categories < 1 || p.Seed.Pets < 1 || p.Seed.Customers < 1:
		return Params{}, fmt.Errorf("params.json: seed sizes must be positive")
	case len(p.Scenarios) == 0:
		return Params{}, fmt.Errorf("params.json: scenarios is empty")
	}
	for _, s := range p.Scenarios {
		if !slices.Contains(Scenarios, s) {
			return Params{}, fmt.Errorf("params.json: unknown scenario %q (known: %v)", s, Scenarios)
		}
	}
	return p, nil
}
