package conform

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// App is the content of an app's app.json: what the suite needs to know that
// the contract leaves to each framework.
type App struct {
	Name          string `json:"name"`
	Admin         *Admin `json:"admin"`
	AdminNAReason string `json:"admin_na_reason"`
}

// ParseAppJSON reads and validates an app.json. Unknown keys are refused so a
// misspelt key cannot silently turn a check into N/A.
func ParseAppJSON(raw []byte) (App, error) {
	var app App
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&app); err != nil {
		return App{}, fmt.Errorf("app.json: %w", err)
	}
	if app.Name == "" {
		return App{}, fmt.Errorf("app.json: name is required")
	}
	if app.Admin == nil {
		if app.AdminNAReason == "" {
			return App{}, fmt.Errorf("app.json: admin is null, so admin_na_reason must say why")
		}
		return app, nil
	}
	a := app.Admin
	if a.LoginPath == "" || a.ListPath == "" || a.UsernameField == "" || a.PasswordField == "" {
		return App{}, fmt.Errorf("app.json: admin needs login_path, list_path, username_field and password_field")
	}
	return app, nil
}
