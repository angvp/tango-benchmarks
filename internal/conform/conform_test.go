package conform

import (
	"context"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/angvp/tango-benchmarks/internal/fakeapp"
)

func run(t *testing.T, faults ...fakeapp.Fault) []Result {
	t.Helper()
	app := fakeapp.New(faults...)
	srv := httptest.NewServer(app.Handler())
	t.Cleanup(srv.Close)
	cfg := Config{BaseURL: srv.URL, Username: fakeapp.Username, Password: fakeapp.Password}
	if !slices.Contains(faults, fakeapp.NoAdmin) {
		cfg.Admin = &Admin{LoginPath: "/admin/login/", ListPath: "/admin/petshop/pet/", UsernameField: "username", PasswordField: "password"}
	} else {
		cfg.AdminNAReason = "this app has no administration interface"
	}
	return Run(context.Background(), cfg)
}

func byID(results []Result, id string) (Result, bool) {
	for _, r := range results {
		if r.ID == id {
			return r, true
		}
	}
	return Result{}, false
}

func TestAConformingAppPassesEveryCheck(t *testing.T) {
	results := run(t)
	if len(results) != 29 {
		t.Fatalf("got %d checks, want 29: the list of checks changed, update the contract page too", len(results))
	}
	for _, r := range results {
		if r.Status != Pass {
			t.Errorf("%s: %s %s", r.ID, r.Status, r.Detail)
		}
	}
}

func TestAnAppWithoutAnAdminIsNotApplicableForAdminChecks(t *testing.T) {
	results := run(t, fakeapp.NoAdmin)
	for _, id := range []string{"admin.protected", "admin.list"} {
		r, ok := byID(results, id)
		if !ok || r.Status != NotApplicable {
			t.Fatalf("%s = %+v, want N/A", id, r)
		}
		if !strings.Contains(r.Detail, "no administration interface") {
			t.Fatalf("%s detail %q does not give the reason", id, r.Detail)
		}
	}
	if Failed(results) {
		t.Fatal("N/A must not count as a failure")
	}
}

// Each fault is a deliberate deviation from the contract. The suite has to
// notice every one of them, and name the check that caught it.
func TestEachDeviationFailsTheMatchingCheck(t *testing.T) {
	cases := []struct {
		fault fakeapp.Fault
		check string
	}{
		{fakeapp.WrongTotal, "list.filter.category"},
		{fakeapp.UnorderedList, "list.default"},
		{fakeapp.PageSizeUncapped, "list.invalid"},
		{fakeapp.IgnoresPriceFilter, "list.filter.price"},
		{fakeapp.IgnoresStatusFilter, "list.filter.status"},
		{fakeapp.NoPastLastPage, "list.beyond_last"},
		{fakeapp.MissingPetIs200, "pet.get_missing"},
		{fakeapp.BadLoginAccepted, "login.bad"},
		{fakeapp.OpenPetCreate, "auth.pet_create_unauth"},
		{fakeapp.OpenPetUpdate, "auth.pet_update_unauth"},
		{fakeapp.OpenOrderCreate, "auth.order_create_unauth"},
		{fakeapp.OpenOrderGet, "auth.order_get_unauth"},
		{fakeapp.CreateIs200, "pet.create"},
		{fakeapp.NoPetValidation, "pet.create_invalid"},
		{fakeapp.UpdateIgnoresBody, "pet.update"},
		{fakeapp.UpdateMissingIs200, "pet.update_missing"},
		{fakeapp.NoUpdateValidation, "pet.update_invalid"},
		{fakeapp.OrderUnknownPetAccepted, "order.create_unknown_pet"},
		{fakeapp.OrderUnknownCustomerAccepted, "order.create_unknown_customer"},
		{fakeapp.OpenAdmin, "admin.protected"},
		{fakeapp.AdminListEmpty, "admin.list"},
	}
	for _, tc := range cases {
		t.Run(string(tc.fault), func(t *testing.T) {
			results := run(t, tc.fault)
			r, ok := byID(results, tc.check)
			if !ok {
				t.Fatalf("no check %s", tc.check)
			}
			if r.Status != Fail {
				t.Fatalf("%s = %s, want fail with fault %s", tc.check, r.Status, tc.fault)
			}
			if !Failed(results) {
				t.Fatal("Failed reports no failure")
			}
		})
	}
}

func TestSeedCountsFailWhenTheDatabaseIsNotFresh(t *testing.T) {
	results := run(t, fakeapp.ExtraPet)
	if r, _ := byID(results, "seed.counts"); r.Status != Fail {
		t.Fatalf("seed.counts = %s, want fail on a database with an extra pet", r.Status)
	}
}

func TestAnUnreachableAppFailsEveryCheckInsteadOfPanicking(t *testing.T) {
	results := Run(context.Background(), Config{BaseURL: "http://127.0.0.1:1", Username: "u", Password: "p", AdminNAReason: "none"})
	if !Failed(results) {
		t.Fatal("an unreachable app must fail")
	}
}

func TestAnAdminThatRateLimitsIsRetriedAndTheRetryIsReported(t *testing.T) {
	defer shortRetryWait()()
	results := run(t, fakeapp.AdminSlowDown)
	r, _ := byID(results, "admin.list")
	if r.Status != Pass {
		t.Fatalf("admin.list = %s %s, want a pass after a retry", r.Status, r.Detail)
	}
	if !strings.Contains(r.Detail, "429") || !strings.Contains(r.Detail, "retried") {
		t.Fatalf("the retry must be reported, got %q", r.Detail)
	}
}

func shortRetryWait() func() {
	old := retryWait
	retryWait = time.Millisecond
	return func() { retryWait = old }
}

func TestAnAdminThatNeverStopsRateLimitingFails(t *testing.T) {
	defer shortRetryWait()()
	results := run(t, fakeapp.AdminAlwaysSlowDown)
	if r, _ := byID(results, "admin.list"); r.Status != Fail {
		t.Fatalf("admin.list = %s, want fail when every attempt is rate limited", r.Status)
	}
}
