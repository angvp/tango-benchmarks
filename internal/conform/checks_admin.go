package conform

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html"
)

const (
	// An admin may rate limit by design (uAdmin allows 3 requests a second by
	// default). A person would not click that fast, so the suite waits and
	// tries again, a few times, and says it did.
	maxAdminRetries = 3
)

// seededPetName matches the name of any seeded pet. The admin may list the
// newest rows first (uAdmin does), so no particular pet is expected.
var seededPetName = regexp.MustCompile(`Pet \d{5}`)

// retryWait is how long to wait after a 429. A variable so tests can shorten it.
var retryWait = 1100 * time.Millisecond

// withRetry runs do until it stops answering 429, at most maxAdminRetries more
// times, and counts the retries.
func withRetry(retries *int, do func() (*http.Response, error)) (*http.Response, error) {
	for attempt := 0; ; attempt++ {
		resp, err := do()
		if err != nil || resp.StatusCode != http.StatusTooManyRequests || attempt == maxAdminRetries {
			return resp, err
		}
		resp.Body.Close()
		*retries++
		time.Sleep(retryWait)
	}
}

var adminChecks = []check{
	{"admin.protected", func(e *env) error {
		if err := adminApplies(e); err != nil {
			return err
		}
		r, err := e.do("GET", e.cfg.Admin.ListPath, false, nil)
		if err != nil {
			return err
		}
		if r.status == http.StatusOK && seededPetName.Match(r.body) {
			return fmt.Errorf("anonymous GET %s shows pet data", e.cfg.Admin.ListPath)
		}
		return nil
	}},
	{"admin.list", func(e *env) error {
		if err := adminApplies(e); err != nil {
			return err
		}
		client := e.jar()
		a := e.cfg.Admin
		loginURL := strings.TrimRight(e.cfg.BaseURL, "/") + a.LoginPath
		retries := 0
		resp, err := withRetry(&retries, func() (*http.Response, error) { return client.Get(loginURL) })
		if err != nil {
			return fmt.Errorf("GET %s: %w", a.LoginPath, err)
		}
		fields, action, err := formFields(resp.Body)
		resp.Body.Close()
		if err != nil {
			return fmt.Errorf("login page: %w", err)
		}
		fields.Set(a.UsernameField, e.cfg.Username)
		fields.Set(a.PasswordField, e.cfg.Password)
		target := loginURL
		if action != "" {
			base, _ := url.Parse(loginURL)
			ref, err := url.Parse(action)
			if err != nil {
				return fmt.Errorf("login form action %q: %w", action, err)
			}
			target = base.ResolveReference(ref).String()
		}
		post, err := withRetry(&retries, func() (*http.Response, error) { return client.PostForm(target, fields) })
		if err != nil {
			return fmt.Errorf("POST login form: %w", err)
		}
		post.Body.Close()
		listURL := strings.TrimRight(e.cfg.BaseURL, "/") + a.ListPath
		list, err := withRetry(&retries, func() (*http.Response, error) { return client.Get(listURL) })
		if err != nil {
			return fmt.Errorf("GET %s: %w", a.ListPath, err)
		}
		defer list.Body.Close()
		body, _ := io.ReadAll(io.LimitReader(list.Body, 8<<20))
		if list.StatusCode != http.StatusOK {
			return fmt.Errorf("signed-in GET %s: status %d, want 200", a.ListPath, list.StatusCode)
		}
		if !strings.Contains(list.Header.Get("Content-Type"), "text/html") {
			return fmt.Errorf("%s is %q, want HTML", a.ListPath, list.Header.Get("Content-Type"))
		}
		if !seededPetName.Match(body) {
			return fmt.Errorf("%s shows no seeded pet name (Pet 00001 to Pet 10000)", a.ListPath)
		}
		if retries > 0 {
			return passWith(fmt.Sprintf("retried after HTTP 429 (%d times): the admin rate limits", retries))
		}
		return nil
	}},
}

func adminApplies(e *env) error {
	if e.cfg.Admin != nil {
		return nil
	}
	if e.cfg.AdminNAReason == "" {
		return fmt.Errorf("no admin configured and no reason given: an app without an admin must say why")
	}
	return notApplicable(e.cfg.AdminNAReason)
}

// formFields returns the named inputs of the first form (hidden ones such as
// CSRF tokens included, as a browser would send them) and its action.
func formFields(r io.Reader) (url.Values, string, error) {
	values := url.Values{}
	z := html.NewTokenizer(r)
	inForm := false
	action := ""
	for {
		switch z.Next() {
		case html.ErrorToken:
			if !inForm {
				return nil, "", fmt.Errorf("no form found")
			}
			return values, action, nil
		case html.StartTagToken, html.SelfClosingTagToken:
			tok := z.Token()
			attrs := map[string]string{}
			for _, a := range tok.Attr {
				attrs[a.Key] = a.Val
			}
			switch tok.Data {
			case "form":
				if inForm {
					return values, action, nil
				}
				inForm, action = true, attrs["action"]
			case "input":
				if inForm && attrs["name"] != "" && attrs["type"] != "submit" {
					values.Set(attrs["name"], attrs["value"])
				}
			}
		case html.EndTagToken:
			if z.Token().Data == "form" && inForm {
				return values, action, nil
			}
		}
	}
}

// FormFields is formFields for other packages: the load runner signs in to an
// admin the same way the suite does.
func FormFields(r io.Reader) (url.Values, string, error) { return formFields(r) }
