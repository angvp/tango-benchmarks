package loadseed

import (
	"strings"
	"testing"
)

const mapJSON = `{
  "tables": [
    {"csv": "categories.csv", "table": "categories", "columns": {"id": "id", "name": "name"}},
    {"csv": "pets.csv", "table": "petshop_pet", "columns": {"id": "id", "name": "title", "price_cents": "price_cents", "status": "status", "category_id": "category_id"}}
  ],
  "truncate_also": ["petshop_order"]
}`

func TestPlanMapsCSVColumnsToTheAppsColumns(t *testing.T) {
	m, err := ParseMap([]byte(mapJSON))
	if err != nil {
		t.Fatal(err)
	}
	stmt, err := m.Tables[1].CopySQL("id,name,price_cents,status,category_id")
	if err != nil {
		t.Fatal(err)
	}
	want := `COPY "petshop_pet" ("id","title","price_cents","status","category_id") FROM STDIN WITH (FORMAT csv, HEADER true)`
	if stmt != want {
		t.Fatalf("got  %s\nwant %s", stmt, want)
	}
}

func TestPlanRefusesAnUnmappedCSVColumn(t *testing.T) {
	m, _ := ParseMap([]byte(mapJSON))
	if _, err := m.Tables[0].CopySQL("id,name,colour"); err == nil || !strings.Contains(err.Error(), "colour") {
		t.Fatalf("an unmapped column must be named in the error, got %v", err)
	}
}

func TestTruncateCoversEveryTableOnce(t *testing.T) {
	m, _ := ParseMap([]byte(mapJSON))
	got := m.TruncateSQL()
	want := `TRUNCATE "categories", "petshop_pet", "petshop_order" RESTART IDENTITY CASCADE`
	if got != want {
		t.Fatalf("got  %s\nwant %s", got, want)
	}
}

func TestSequenceResetUsesTheMappedIDColumn(t *testing.T) {
	m, _ := ParseMap([]byte(mapJSON))
	got := m.Tables[1].SetvalSQL()
	want := `SELECT setval(pg_get_serial_sequence('"petshop_pet"', 'id'), (SELECT max("id") FROM "petshop_pet"))`
	if got != want {
		t.Fatalf("got  %s\nwant %s", got, want)
	}
}

func TestMapRejectsAnIdentifierThatIsNotPlain(t *testing.T) {
	bad := `{"tables":[{"csv":"c.csv","table":"x; DROP TABLE y","columns":{"id":"id"}}]}`
	if _, err := ParseMap([]byte(bad)); err == nil {
		t.Fatal("a table name that is not a plain identifier must be refused")
	}
}

func TestMapNeedsTheIDColumnMapped(t *testing.T) {
	bad := `{"tables":[{"csv":"c.csv","table":"x","columns":{"name":"name"}}]}`
	if _, err := ParseMap([]byte(bad)); err == nil {
		t.Fatal("a table whose id column is unmapped cannot have its sequence reset")
	}
}
