package load

import (
	"strings"
	"testing"
)

func rec(scenario string, run int, rps float64) Record {
	return Record{App: "x", Run: run, PeakMemMB: 100, ColdStartMs: 500,
		Result: Result{Scenario: scenario, Stamp: Stamp, Requests: 100, OK: 100, RPS: rps, P50Ms: 1, P95Ms: 2, P99Ms: 3, Statuses: map[string]int{"200": 100}}}
}

func TestSummarizeGivesMedianAndSpreadPerScenario(t *testing.T) {
	var recs []Record
	for i, v := range []float64{100, 90, 110, 80, 120} {
		recs = append(recs, rec("pet", i+1, v), rec("list", i+1, 50))
	}
	sum, err := Summarize(recs)
	if err != nil {
		t.Fatal(err)
	}
	if len(sum.Scenarios) != 2 || sum.Scenarios[0].Scenario != "list" || sum.Scenarios[1].Scenario != "pet" {
		t.Fatalf("scenarios out of order: %+v", sum.Scenarios)
	}
	pet := sum.Scenarios[1]
	if pet.RPS.Median != 100 || pet.RPS.Min != 80 || pet.RPS.Max != 120 || pet.RPS.RangePct != 40 {
		t.Errorf("rps spread = %+v", pet.RPS)
	}
	table := sum.Table()
	if !strings.Contains(table, Stamp) || !strings.Contains(table, "range 40.0%") {
		t.Errorf("table lacks the stamp or the spread:\n%s", table)
	}
}

func TestSummarizeRefusesTooFewRunsAndMixedRuns(t *testing.T) {
	if _, err := Summarize([]Record{rec("pet", 1, 1), rec("pet", 2, 1)}); err == nil || !strings.Contains(err.Error(), "at least 5") {
		t.Errorf("err = %v", err)
	}
	other := rec("pet", 1, 1)
	other.App = "y"
	if _, err := Summarize([]Record{rec("pet", 1, 1), other}); err == nil {
		t.Error("mixed apps must fail")
	}
	if _, err := Summarize(nil); err == nil {
		t.Error("no records must fail")
	}
}

func TestSummarizeKeepsNotApplicableAndCountsFailureStatuses(t *testing.T) {
	na := Record{App: "x", Run: 1, Result: Result{Scenario: "admin", Stamp: Stamp, NotApplicable: true, NAReason: "no admin"}}
	sum, err := Summarize([]Record{na})
	if err != nil || !sum.Scenarios[0].NotApplicable {
		t.Fatalf("%+v %v", sum, err)
	}
	if !strings.Contains(sum.Table(), "N/A: no admin") {
		t.Errorf("table:\n%s", sum.Table())
	}
	var recs []Record
	for i := 1; i <= 5; i++ {
		r := rec("admin", i, 10)
		r.Requests, r.OK, r.Errors = 100, 40, 60
		r.Statuses = map[string]int{"200": 40, "429": 60}
		recs = append(recs, r)
	}
	sum, _ = Summarize(recs)
	if got := sum.Scenarios[0]; got.Statuses["429"] != 300 || got.ErrorPercent != 60 {
		t.Errorf("%+v", got)
	}
}
