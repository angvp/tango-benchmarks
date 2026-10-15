// Package conform is the black-box conformance suite for contract/petshop.md.
// It talks to an app only over HTTP, and records each check as pass, fail or
// N/A. A framework gets N/A with a reason, never a silent skip.
package conform

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"
)

type Status string

const (
	Pass          Status = "pass"
	Fail          Status = "fail"
	NotApplicable Status = "n/a"
)

type Result struct {
	ID     string `json:"id"`
	Status Status `json:"status"`
	Detail string `json:"detail,omitempty"`
}

// Admin describes an app's HTML administration interface (see the contract).
type Admin struct {
	LoginPath     string `json:"login_path"`
	ListPath      string `json:"list_path"`
	UsernameField string `json:"username_field"`
	PasswordField string `json:"password_field"`
}

type Config struct {
	BaseURL  string
	Username string
	Password string
	// Admin is nil for an app with no administration interface; then
	// AdminNAReason must say why.
	Admin         *Admin
	AdminNAReason string
}

// Failed reports whether any check failed. N/A is not a failure.
func Failed(results []Result) bool {
	for _, r := range results {
		if r.Status == Fail {
			return true
		}
	}
	return false
}

type notApplicable string

func (n notApplicable) Error() string { return string(n) }

// env is the shared state of one run: checks run in order and later ones use
// what earlier ones created.
type env struct {
	cfg     Config
	http    *http.Client
	token   string
	petID   int
	orderID int
}

type check struct {
	id  string
	run func(*env) error
}

func Run(ctx context.Context, cfg Config) []Result {
	e := &env{cfg: cfg, http: &http.Client{Timeout: 30 * time.Second, CheckRedirect: noRedirect}}
	var checks []check
	checks = append(checks, listChecks...)
	checks = append(checks, authChecks...)
	checks = append(checks, writeChecks...)
	checks = append(checks, adminChecks...)
	results := make([]Result, 0, len(checks))
	for _, c := range checks {
		if err := ctx.Err(); err != nil {
			results = append(results, Result{c.id, Fail, err.Error()})
			continue
		}
		results = append(results, outcome(c.id, c.run(e)))
	}
	return results
}

// passWith is returned by a check that passed but has something to report.
type passWith string

func (p passWith) Error() string { return string(p) }

func outcome(id string, err error) Result {
	if err == nil {
		return Result{ID: id, Status: Pass}
	}
	if note, ok := err.(passWith); ok {
		return Result{ID: id, Status: Pass, Detail: string(note)}
	}
	if na, ok := err.(notApplicable); ok {
		return Result{ID: id, Status: NotApplicable, Detail: string(na)}
	}
	return Result{ID: id, Status: Fail, Detail: err.Error()}
}

func noRedirect(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }

type response struct {
	status int
	header http.Header
	body   []byte
}

func (e *env) do(method, path string, token bool, body any) (response, error) {
	return e.doWith(e.http, method, path, token, body)
}

func (e *env) doWith(client *http.Client, method, path string, token bool, body any) (response, error) {
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return response{}, err
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, strings.TrimRight(e.cfg.BaseURL, "/")+path, reader)
	if err != nil {
		return response{}, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token && e.token != "" {
		req.Header.Set("Authorization", "Bearer "+e.token)
	}
	resp, err := client.Do(req)
	if err != nil {
		return response{}, fmt.Errorf("%s %s: %w", method, path, err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return response{}, fmt.Errorf("%s %s: reading body: %w", method, path, err)
	}
	return response{status: resp.StatusCode, header: resp.Header, body: raw}, nil
}

func (e *env) jar() *http.Client {
	jar, _ := cookiejar.New(nil)
	return &http.Client{Timeout: 30 * time.Second, Jar: jar}
}

// expect sends a request and requires the status; it returns the response.
func (e *env) expect(want int, method, path string, token bool, body any) (response, error) {
	r, err := e.do(method, path, token, body)
	if err != nil {
		return r, err
	}
	if r.status != want {
		return r, fmt.Errorf("%s %s: status %d, want %d", method, path, r.status, want)
	}
	return r, nil
}

type pet struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	PriceCents int    `json:"price_cents"`
	Status     string `json:"status"`
	CategoryID int    `json:"category_id"`
}

type page struct {
	Items    []pet `json:"items"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Total    int   `json:"total"`
}

func decode(r response, into any) error {
	if err := json.Unmarshal(r.body, into); err != nil {
		return fmt.Errorf("response is not the expected JSON: %w", err)
	}
	return nil
}

func (e *env) list(query string) (page, error) {
	r, err := e.expect(http.StatusOK, "GET", "/api/pets"+query, false, nil)
	if err != nil {
		return page{}, err
	}
	var p page
	return p, decode(r, &p)
}
