package load

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/angvp/tango-benchmarks/internal/conform"
	"github.com/angvp/tango-benchmarks/internal/fakeapp"
)

func testParams() Params {
	return Params{WarmupSeconds: 0, DurationSeconds: 1, Concurrency: 4, Runs: 5,
		Seed: SeedShape{Categories: 20, Pets: 10000, Customers: 1000}, ListPageSize: 20, ReadPercent: 90}
}

func target(t *testing.T) Target {
	t.Helper()
	srv := httptest.NewServer(fakeapp.New().Handler())
	t.Cleanup(srv.Close)
	return Target{BaseURL: srv.URL, Username: fakeapp.Username, Password: fakeapp.Password,
		Admin: &conform.Admin{LoginPath: "/admin/login/", ListPath: "/admin/petshop/pet/", UsernameField: "username", PasswordField: "password"}}
}

func TestEveryScenarioRunsCleanAgainstAConformingApp(t *testing.T) {
	tg := target(t)
	for _, name := range []string{"list", "pet", "order", "admin", "mix"} {
		res, err := Run(context.Background(), testParams(), tg, name)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if res.Requests == 0 || res.Errors != 0 || res.OK != res.Requests {
			t.Errorf("%s: %+v", name, res)
		}
		if res.P50Ms <= 0 || res.P99Ms < res.P95Ms || res.P95Ms < res.P50Ms || res.RPS <= 0 {
			t.Errorf("%s: implausible latencies %+v", name, res)
		}
		if res.Stamp != Stamp {
			t.Errorf("%s: stamp = %q", name, res.Stamp)
		}
	}
}

func TestAnAppWithoutAnAdminIsNotApplicableForTheAdminScenario(t *testing.T) {
	tg := target(t)
	tg.Admin, tg.AdminNAReason = nil, "no admin by design"
	res, err := Run(context.Background(), testParams(), tg, "admin")
	if err != nil {
		t.Fatal(err)
	}
	if !res.NotApplicable || res.NAReason != "no admin by design" || res.Requests != 0 {
		t.Errorf("got %+v", res)
	}
}

func TestErrorsAreCountedByStatusAndKeptOutOfLatency(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/login" {
			w.Write([]byte(`{"token":"t"}`))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()
	res, err := Run(context.Background(), testParams(), Target{BaseURL: srv.URL, Username: "u", Password: "p"}, "pet")
	if err != nil {
		t.Fatal(err)
	}
	if res.OK != 0 || res.Errors != res.Requests || res.Statuses["500"] != res.Requests {
		t.Errorf("got %+v", res)
	}
}

func TestAnUnknownScenarioOrAnUnreachableAppIsAnError(t *testing.T) {
	if _, err := Run(context.Background(), testParams(), target(t), "nope"); err == nil {
		t.Error("unknown scenario must fail")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if _, err := Run(ctx, testParams(), Target{BaseURL: "http://127.0.0.1:1", Username: "u", Password: "p"}, "pet"); err == nil {
		t.Error("an unreachable app must fail, not report zero requests")
	}
}
