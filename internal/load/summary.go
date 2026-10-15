package load

import (
	"fmt"
	"slices"
	"strings"
)

// Record is one run of one scenario with the measurements around it.
type Record struct {
	App         string  `json:"app"`
	Run         int     `json:"run"`
	ColdStartMs float64 `json:"cold_start_ms,omitempty"`
	PeakMemMB   float64 `json:"peak_mem_mb,omitempty"`
	Result
}

// ScenarioSummary is every run of one scenario, reduced to medians and
// spreads.
type ScenarioSummary struct {
	Scenario      string  `json:"scenario"`
	NotApplicable bool    `json:"not_applicable,omitempty"`
	NAReason      string  `json:"na_reason,omitempty"`
	Runs          int     `json:"runs"`
	RPS           Spread  `json:"rps"`
	P50Ms         Spread  `json:"p50_ms"`
	P95Ms         Spread  `json:"p95_ms"`
	P99Ms         Spread  `json:"p99_ms"`
	PeakMemMB     Spread  `json:"peak_mem_mb"`
	ColdStartMs   Spread  `json:"cold_start_ms"`
	ErrorPercent  float64 `json:"error_percent"`
	// Statuses totals the non-success statuses over all runs.
	Statuses map[string]int `json:"error_statuses,omitempty"`
}

// Summary is what a set of runs of one app says. It keeps the stamp of its
// runs.
type Summary struct {
	App       string            `json:"app"`
	Stamp     string            `json:"stamp"`
	Scenarios []ScenarioSummary `json:"scenarios"`
}

// Summarize groups records by scenario, in the order of Scenarios. It fails
// when a scenario has fewer than MinRuns runs, or the records mix apps or
// stamps.
func Summarize(records []Record) (Summary, error) {
	if len(records) == 0 {
		return Summary{}, fmt.Errorf("no records")
	}
	sum := Summary{App: records[0].App, Stamp: records[0].Stamp}
	for _, r := range records {
		if r.App != sum.App || r.Stamp != sum.Stamp {
			return Summary{}, fmt.Errorf("records mix apps or stamps: %s/%q and %s/%q", sum.App, sum.Stamp, r.App, r.Stamp)
		}
	}
	for _, name := range Scenarios {
		var runs []Record
		for _, r := range records {
			if r.Scenario == name {
				runs = append(runs, r)
			}
		}
		if len(runs) == 0 {
			continue
		}
		s, err := summarizeScenario(name, runs)
		if err != nil {
			return Summary{}, err
		}
		sum.Scenarios = append(sum.Scenarios, s)
	}
	return sum, nil
}

func summarizeScenario(name string, runs []Record) (ScenarioSummary, error) {
	s := ScenarioSummary{Scenario: name, Runs: len(runs)}
	if runs[0].NotApplicable {
		s.NotApplicable, s.NAReason = true, runs[0].NAReason
		return s, nil
	}
	if len(runs) < MinRuns {
		return s, fmt.Errorf("%s: %d runs, at least %d are needed", name, len(runs), MinRuns)
	}
	column := func(f func(Record) float64) Spread {
		vals := make([]float64, 0, len(runs))
		for _, r := range runs {
			vals = append(vals, f(r))
		}
		sp, _ := SpreadOf(vals)
		return sp
	}
	s.RPS = column(func(r Record) float64 { return r.RPS })
	s.P50Ms = column(func(r Record) float64 { return r.P50Ms })
	s.P95Ms = column(func(r Record) float64 { return r.P95Ms })
	s.P99Ms = column(func(r Record) float64 { return r.P99Ms })
	s.PeakMemMB = column(func(r Record) float64 { return r.PeakMemMB })
	s.ColdStartMs = column(func(r Record) float64 { return r.ColdStartMs })
	var reqs, errs int
	for _, r := range runs {
		reqs += r.Requests
		errs += r.Errors
		for status, n := range r.Statuses {
			if status != "200" && status != "201" {
				if s.Statuses == nil {
					s.Statuses = map[string]int{}
				}
				s.Statuses[status] += n
			}
		}
	}
	if reqs > 0 {
		s.ErrorPercent = float64(errs) / float64(reqs) * 100
	}
	return s, nil
}

// Table renders a summary for a terminal. Every scenario shows the spread of
// its runs next to the median, and the stamp heads the table.
func (s Summary) Table() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s [%s]\n", s.App, s.Stamp)
	fmt.Fprintf(&b, "%-7s %5s %18s %10s %10s %10s %9s %9s %7s\n", "scenario", "runs", "req/s (min-max)", "p50 ms", "p95 ms", "p99 ms", "mem MB", "cold ms", "err %")
	for _, sc := range s.Scenarios {
		if sc.NotApplicable {
			fmt.Fprintf(&b, "%-7s  N/A: %s\n", sc.Scenario, sc.NAReason)
			continue
		}
		fmt.Fprintf(&b, "%-7s %5d %8.0f (%.0f-%.0f) %10.2f %10.2f %10.2f %9.0f %9.0f %7.2f\n", sc.Scenario, sc.Runs,
			sc.RPS.Median, sc.RPS.Min, sc.RPS.Max, sc.P50Ms.Median, sc.P95Ms.Median, sc.P99Ms.Median,
			sc.PeakMemMB.Median, sc.ColdStartMs.Median, sc.ErrorPercent)
		if len(sc.Statuses) > 0 {
			keys := slices.Sorted(func(yield func(string) bool) {
				for k := range sc.Statuses {
					if !yield(k) {
						return
					}
				}
			})
			for _, k := range keys {
				fmt.Fprintf(&b, "        non-success HTTP %s: %d\n", k, sc.Statuses[k])
			}
		}
		fmt.Fprintf(&b, "        req/s run-to-run range %.1f%% of the median\n", sc.RPS.RangePct)
	}
	return b.String()
}
