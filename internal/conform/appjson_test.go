package conform

import "testing"

func TestParseAppJSON(t *testing.T) {
	app, err := ParseAppJSON([]byte(`{"name":"django","admin":{"login_path":"/admin/login/","list_path":"/admin/p/","username_field":"username","password_field":"password"}}`))
	if err != nil || app.Admin == nil || app.Admin.ListPath != "/admin/p/" || app.Name != "django" {
		t.Fatalf("with admin: %+v, %v", app, err)
	}
	app, err = ParseAppJSON([]byte(`{"name":"baseline","admin":null,"admin_na_reason":"the baseline is a plain HTTP server"}`))
	if err != nil || app.Admin != nil || app.AdminNAReason == "" {
		t.Fatalf("without admin: %+v, %v", app, err)
	}
}

func TestAppJSONWithoutAdminMustGiveAReason(t *testing.T) {
	if _, err := ParseAppJSON([]byte(`{"name":"x","admin":null}`)); err == nil {
		t.Fatal("an app with no admin and no reason must be refused")
	}
}

func TestAppJSONNeedsAName(t *testing.T) {
	if _, err := ParseAppJSON([]byte(`{"admin":null,"admin_na_reason":"r"}`)); err == nil {
		t.Fatal("an app.json without a name must be refused")
	}
}

func TestAppJSONAdminNeedsAllFields(t *testing.T) {
	if _, err := ParseAppJSON([]byte(`{"name":"x","admin":{"login_path":"/l/"}}`)); err == nil {
		t.Fatal("a partial admin block must be refused")
	}
}

func TestAppJSONRejectsUnknownKeys(t *testing.T) {
	if _, err := ParseAppJSON([]byte(`{"name":"x","admin":null,"admin_na_reason":"r","adminn":1}`)); err == nil {
		t.Fatal("a misspelt key must be refused, not ignored")
	}
}
