package conform

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

const firstPetName = "Pet 00001"

var adminChecks = []check{
	{"admin.protected", func(e *env) error {
		if err := adminApplies(e); err != nil {
			return err
		}
		r, err := e.do("GET", e.cfg.Admin.ListPath, false, nil)
		if err != nil {
			return err
		}
		if r.status == http.StatusOK && strings.Contains(string(r.body), firstPetName) {
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
		resp, err := client.Get(loginURL)
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
		post, err := client.PostForm(target, fields)
		if err != nil {
			return fmt.Errorf("POST login form: %w", err)
		}
		post.Body.Close()
		list, err := client.Get(strings.TrimRight(e.cfg.BaseURL, "/") + a.ListPath)
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
		if !strings.Contains(string(body), firstPetName) {
			return fmt.Errorf("%s does not show %q", a.ListPath, firstPetName)
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
