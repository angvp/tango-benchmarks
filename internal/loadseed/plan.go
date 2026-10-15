// Package loadseed turns an app's seed-map.json into the SQL that loads the
// shared CSV files into the tables the app created. The SQL is built from
// validated identifiers, never from free text.
package loadseed

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

type Map struct {
	Tables       []Table  `json:"tables"`
	TruncateAlso []string `json:"truncate_also"`
}

type Table struct {
	CSV     string            `json:"csv"`
	Table   string            `json:"table"`
	Columns map[string]string `json:"columns"` // CSV column -> database column
}

var plainIdentifier = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func ParseMap(raw []byte) (Map, error) {
	var m Map
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&m); err != nil {
		return Map{}, fmt.Errorf("seed-map.json: %w", err)
	}
	if len(m.Tables) == 0 {
		return Map{}, fmt.Errorf("seed-map.json: no tables")
	}
	for _, t := range m.Tables {
		if !plainIdentifier.MatchString(t.Table) {
			return Map{}, fmt.Errorf("seed-map.json: table %q is not a plain identifier", t.Table)
		}
		if _, ok := t.Columns["id"]; !ok {
			return Map{}, fmt.Errorf("seed-map.json: table %s does not map the id column", t.Table)
		}
		for csvCol, dbCol := range t.Columns {
			if !plainIdentifier.MatchString(csvCol) || !plainIdentifier.MatchString(dbCol) {
				return Map{}, fmt.Errorf("seed-map.json: column %q -> %q is not a plain identifier", csvCol, dbCol)
			}
		}
	}
	for _, name := range m.TruncateAlso {
		if !plainIdentifier.MatchString(name) {
			return Map{}, fmt.Errorf("seed-map.json: truncate_also %q is not a plain identifier", name)
		}
	}
	return m, nil
}

func quote(name string) string { return `"` + name + `"` }

// CopySQL is the COPY statement for a CSV whose header line is header. The
// columns are listed in CSV order, translated to the app's own column names.
func (t Table) CopySQL(header string) (string, error) {
	var cols []string
	for _, csvCol := range strings.Split(strings.TrimSpace(header), ",") {
		dbCol, ok := t.Columns[csvCol]
		if !ok {
			return "", fmt.Errorf("%s: CSV column %q is not mapped for table %s", t.CSV, csvCol, t.Table)
		}
		cols = append(cols, quote(dbCol))
	}
	return fmt.Sprintf("COPY %s (%s) FROM STDIN WITH (FORMAT csv, HEADER true)", quote(t.Table), strings.Join(cols, ",")), nil
}

// SetvalSQL moves the id sequence past the loaded rows, so a row created
// afterwards by the app gets a fresh id. COPY does not advance it.
func (t Table) SetvalSQL() string {
	id := t.Columns["id"]
	return fmt.Sprintf(`SELECT setval(pg_get_serial_sequence('%s', '%s'), (SELECT max(%s) FROM %s))`,
		quote(t.Table), id, quote(id), quote(t.Table))
}

func (m Map) TruncateSQL() string {
	var names []string
	for _, t := range m.Tables {
		names = append(names, quote(t.Table))
	}
	for _, n := range m.TruncateAlso {
		names = append(names, quote(n))
	}
	return "TRUNCATE " + strings.Join(names, ", ") + " RESTART IDENTITY CASCADE"
}
