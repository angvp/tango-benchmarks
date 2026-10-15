package seed

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// The committed files under seed/data are what every app is loaded with, so
// they must be exactly what the generator writes.
func TestCommittedCSVMatchesGenerator(t *testing.T) {
	dir := t.TempDir()
	if err := WriteCSV(dir); err != nil {
		t.Fatalf("WriteCSV: %v", err)
	}
	for _, name := range []string{"categories.csv", "pets.csv", "customers.csv"} {
		want, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile(filepath.Join("..", "..", "seed", "data", name))
		if err != nil {
			t.Fatalf("%s is not committed: %v (run: go run ./cmd/genseed)", name, err)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("seed/data/%s differs from the generator (run: go run ./cmd/genseed)", name)
		}
	}
}

func TestCSVHeadersAndRowCounts(t *testing.T) {
	dir := t.TempDir()
	if err := WriteCSV(dir); err != nil {
		t.Fatal(err)
	}
	pets, _ := os.ReadFile(filepath.Join(dir, "pets.csv"))
	lines := bytes.Split(bytes.TrimSpace(pets), []byte("\n"))
	if string(lines[0]) != "id,name,price_cents,status,category_id" {
		t.Fatalf("pets header = %q", lines[0])
	}
	if len(lines) != PetCount+1 {
		t.Fatalf("pets.csv has %d lines, want %d", len(lines), PetCount+1)
	}
	if string(lines[1]) != "1,Pet 00001,8419,available,1" {
		t.Fatalf("first pet row = %q", lines[1])
	}
}
