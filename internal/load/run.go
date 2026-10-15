package load

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/angvp/tango-benchmarks/internal/conform"
)

// Target is the app under load.
type Target struct {
	BaseURL       string
	Username      string
	Password      string
	Admin         *conform.Admin
	AdminNAReason string
}

// Result is one run of one scenario.
type Result struct {
	Scenario      string         `json:"scenario"`
	Stamp         string         `json:"stamp"`
	NotApplicable bool           `json:"not_applicable,omitempty"`
	NAReason      string         `json:"na_reason,omitempty"`
	Concurrency   int            `json:"concurrency"`
	Seconds       float64        `json:"seconds"`
	Requests      int            `json:"requests"`
	OK            int            `json:"ok"`
	Errors        int            `json:"errors"`
	Statuses      map[string]int `json:"statuses"`
	RPS           float64        `json:"rps"`
	P50Ms         float64        `json:"p50_ms"`
	P95Ms         float64        `json:"p95_ms"`
	P99Ms         float64        `json:"p99_ms"`
	// FirstError is the first transport failure (status 0), if any.
	FirstError string `json:"first_error,omitempty"`
}

// request builds the next request of a scenario and says which status counts
// as success.
type request func(rng *rand.Rand) (*http.Request, int, error)

// Run drives one scenario: warm-up, then a timed measurement with
// p.Concurrency workers in a closed loop. Only successful requests count in
// the latency percentiles; failures are counted by status.
func Run(ctx context.Context, p Params, t Target, scenario string) (Result, error) {
	if !slices.Contains(Scenarios, scenario) {
		return Result{}, fmt.Errorf("unknown scenario %q", scenario)
	}
	base := Result{Scenario: scenario, Stamp: Stamp, Concurrency: p.Concurrency}
	if scenario == "admin" && t.Admin == nil {
		base.NotApplicable, base.NAReason = true, t.AdminNAReason
		return base, nil
	}
	client := &http.Client{Transport: &http.Transport{MaxIdleConnsPerHost: p.Concurrency}, Timeout: 30 * time.Second}
	next, err := build(p, t, scenario, client)
	if err != nil {
		return Result{}, err
	}
	if err := probe(next, client); err != nil {
		return Result{}, err
	}
	if p.WarmupSeconds > 0 {
		drive(ctx, p, client, next, time.Duration(p.WarmupSeconds)*time.Second)
	}
	start := time.Now()
	samples := drive(ctx, p, client, next, time.Duration(p.DurationSeconds)*time.Second)
	return summarize(base, samples, time.Since(start)), nil
}

type sample struct {
	latency time.Duration
	status  int // 0 for a transport failure
	err     string
	ok      bool
}

func probe(next request, client *http.Client) error {
	req, _, err := next(rand.New(rand.NewPCG(1, 1)))
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("the app is not reachable: %w", err)
	}
	io.Copy(io.Discard, resp.Body)
	return resp.Body.Close()
}

func drive(ctx context.Context, p Params, client *http.Client, next request, d time.Duration) []sample {
	ctx, cancel := context.WithTimeout(ctx, d)
	defer cancel()
	out := make([][]sample, p.Concurrency)
	var wg sync.WaitGroup
	for w := range p.Concurrency {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rng := rand.New(rand.NewPCG(42, uint64(w)))
			for ctx.Err() == nil {
				req, want, err := next(rng)
				if err != nil {
					return
				}
				began := time.Now()
				resp, err := client.Do(req.WithContext(ctx))
				if err != nil {
					if ctx.Err() == nil {
						out[w] = append(out[w], sample{status: 0, err: err.Error()})
					}
					continue
				}
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
				out[w] = append(out[w], sample{latency: time.Since(began), status: resp.StatusCode, ok: resp.StatusCode == want})
			}
		}()
	}
	wg.Wait()
	return slices.Concat(out...)
}

func summarize(res Result, samples []sample, elapsed time.Duration) Result {
	res.Statuses = map[string]int{}
	var lat []time.Duration
	for _, s := range samples {
		res.Requests++
		if s.err != "" && res.FirstError == "" {
			res.FirstError = s.err
		}
		res.Statuses[strconv.Itoa(s.status)]++
		if s.ok {
			res.OK++
			lat = append(lat, s.latency)
		} else {
			res.Errors++
		}
	}
	slices.Sort(lat)
	res.Seconds = elapsed.Seconds()
	if res.Seconds > 0 {
		res.RPS = float64(res.OK) / res.Seconds
	}
	ms := func(d time.Duration) float64 { return float64(d) / float64(time.Millisecond) }
	res.P50Ms, res.P95Ms, res.P99Ms = ms(Percentile(lat, 50)), ms(Percentile(lat, 95)), ms(Percentile(lat, 99))
	return res
}

func build(p Params, t Target, scenario string, client *http.Client) (request, error) {
	base := strings.TrimRight(t.BaseURL, "/")
	list := func(rng *rand.Rand) (*http.Request, int, error) {
		q := url.Values{"category": {strconv.Itoa(1 + rng.IntN(p.Seed.Categories))}, "status": {"available"},
			"page": {strconv.Itoa(1 + rng.IntN(5))}, "page_size": {strconv.Itoa(p.ListPageSize)}}
		req, err := http.NewRequest("GET", base+"/api/pets?"+q.Encode(), nil)
		return req, http.StatusOK, err
	}
	pet := func(rng *rand.Rand) (*http.Request, int, error) {
		req, err := http.NewRequest("GET", base+"/api/pets/"+strconv.Itoa(1+rng.IntN(p.Seed.Pets)), nil)
		return req, http.StatusOK, err
	}
	switch scenario {
	case "list":
		return list, nil
	case "pet":
		return pet, nil
	case "admin":
		return adminList(base, t)
	}
	token, err := login(client, base, t)
	if err != nil {
		return nil, err
	}
	order := func(rng *rand.Rand) (*http.Request, int, error) {
		body, _ := json.Marshal(map[string]int{"pet_id": 1 + rng.IntN(p.Seed.Pets), "customer_id": 1 + rng.IntN(p.Seed.Customers)})
		req, err := http.NewRequest("POST", base+"/api/orders", bytes.NewReader(body))
		if err != nil {
			return nil, 0, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		return req, http.StatusCreated, nil
	}
	if scenario == "order" {
		return order, nil
	}
	return func(rng *rand.Rand) (*http.Request, int, error) { // mix
		switch roll := rng.IntN(100); {
		case roll >= p.ReadPercent:
			return order(rng)
		case roll%2 == 0:
			return list(rng)
		default:
			return pet(rng)
		}
	}, nil
}

func login(client *http.Client, base string, t Target) (string, error) {
	body, _ := json.Marshal(map[string]string{"username": t.Username, "password": t.Password})
	resp, err := client.Post(base+"/api/login", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("POST /api/login: %w", err)
	}
	defer resp.Body.Close()
	var out struct {
		Token string `json:"token"`
	}
	if resp.StatusCode != http.StatusOK || json.NewDecoder(resp.Body).Decode(&out) != nil || out.Token == "" {
		return "", fmt.Errorf("POST /api/login: status %d, no token", resp.StatusCode)
	}
	return out.Token, nil
}

// adminList signs in through the admin's HTML form once, then repeats the
// list page with the session cookie.
func adminList(base string, t Target) (request, error) {
	jar, _ := cookiejar.New(nil)
	// Sign in the way a person does: one step at a time, not following the
	// redirect after the login (it would spend another request), pausing
	// between steps so a rate-limited admin (uAdmin allows 3 requests a
	// second) lets each one through.
	client := &http.Client{Jar: jar, Timeout: 30 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	a := t.Admin
	loginURL := base + a.LoginPath
	resp, err := retry429(func() (*http.Response, error) { return client.Get(loginURL) })
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", a.LoginPath, err)
	}
	fields, action, err := conform.FormFields(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("admin login page: %w", err)
	}
	fields.Set(a.UsernameField, t.Username)
	fields.Set(a.PasswordField, t.Password)
	time.Sleep(adminStepPause)
	target := loginURL
	if action != "" {
		ref, err := url.Parse(action)
		if err != nil {
			return nil, fmt.Errorf("admin login form action %q: %w", action, err)
		}
		u, _ := url.Parse(loginURL)
		target = u.ResolveReference(ref).String()
	}
	post, err := retry429(func() (*http.Response, error) { return client.PostForm(target, fields) })
	if err != nil {
		return nil, fmt.Errorf("POST admin login: %w", err)
	}
	post.Body.Close()
	time.Sleep(adminStepPause)
	listURL := base + a.ListPath
	// Signed in for real? One list request must show a seeded pet before the
	// measurement starts, or every request after it would be measuring a login page.
	check, err := retry429(func() (*http.Response, error) { return client.Get(listURL) })
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", a.ListPath, err)
	}
	body, _ := io.ReadAll(io.LimitReader(check.Body, 8<<20))
	check.Body.Close()
	if check.StatusCode != http.StatusOK || !seededPet.Match(body) {
		return nil, fmt.Errorf("admin sign-in failed: GET %s answered %d with no seeded pet name", a.ListPath, check.StatusCode)
	}
	cookies := jar.Cookies(mustParse(listURL))
	return func(*rand.Rand) (*http.Request, int, error) {
		req, err := http.NewRequest("GET", listURL, nil)
		if err != nil {
			return nil, 0, err
		}
		for _, c := range cookies {
			req.AddCookie(c)
		}
		return req, http.StatusOK, nil
	}, nil
}

// adminStepPause separates the steps of an admin sign-in.
var adminStepPause = 2 * time.Second

var seededPet = regexp.MustCompile(`Pet \d{5}`)

// retry429 repeats a request while it answers 429, as a person would after
// a pause: an admin that rate limits is still signed in to.
func retry429(do func() (*http.Response, error)) (*http.Response, error) {
	for attempt := 0; ; attempt++ {
		resp, err := do()
		if err != nil || resp.StatusCode != http.StatusTooManyRequests || attempt == 5 {
			return resp, err
		}
		resp.Body.Close()
		time.Sleep(2200 * time.Millisecond)
	}
}

func mustParse(raw string) *url.URL {
	u, err := url.Parse(raw)
	if err != nil {
		panic(err)
	}
	return u
}
