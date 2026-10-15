package seed

import "testing"

// The expected numbers are worked out by hand from the rules in
// contract/petshop.md, not recomputed with the generator's own formulas:
// 10,000 pets spread over 20 categories is 500 each; the status residues
// 0-13 (14 of 20) are available, 14-16 pending, 17-19 sold.

func TestCountsAreFixed(t *testing.T) {
	if got := len(Categories()); got != 20 {
		t.Fatalf("categories = %d, want 20", got)
	}
	if got := len(Pets()); got != 10000 {
		t.Fatalf("pets = %d, want 10000", got)
	}
	if got := len(Customers()); got != 1000 {
		t.Fatalf("customers = %d, want 1000", got)
	}
}

func TestPetsSpreadEvenly(t *testing.T) {
	perCategory := map[int]int{}
	perStatus := map[string]int{}
	for _, p := range Pets() {
		perCategory[p.CategoryID]++
		perStatus[p.Status]++
	}
	for c := 1; c <= 20; c++ {
		if perCategory[c] != 500 {
			t.Fatalf("category %d has %d pets, want 500", c, perCategory[c])
		}
	}
	want := map[string]int{"available": 7000, "pending": 1500, "sold": 1500}
	for status, n := range want {
		if perStatus[status] != n {
			t.Fatalf("status %s has %d pets, want %d", status, perStatus[status], n)
		}
	}
}

func TestKnownRows(t *testing.T) {
	pets := Pets()
	// id 1: residue 1 -> available; category (1-1)%20+1 = 1; price 500 + 7919%199500.
	first := pets[0]
	if first.ID != 1 || first.Name != "Pet 00001" || first.Status != "available" || first.CategoryID != 1 || first.PriceCents != 8419 {
		t.Fatalf("pet 1 = %+v", first)
	}
	// id 15: residue 15 -> pending; category 15.
	if p := pets[14]; p.Status != "pending" || p.CategoryID != 15 {
		t.Fatalf("pet 15 = %+v", p)
	}
	// id 20: residue 0 -> available; category 20.
	if p := pets[19]; p.Status != "available" || p.CategoryID != 20 {
		t.Fatalf("pet 20 = %+v", p)
	}
	// id 18: residue 18 -> sold.
	if p := pets[17]; p.Status != "sold" {
		t.Fatalf("pet 18 = %+v", p)
	}
	if c := Customers()[41]; c.Name != "Customer 0042" || c.Email != "customer0042@example.test" {
		t.Fatalf("customer 42 = %+v", c)
	}
	if c := Categories()[2]; c.ID != 3 || c.Name != "Category 03" {
		t.Fatalf("category 3 = %+v", c)
	}
}

func TestPricesStayInRange(t *testing.T) {
	for _, p := range Pets() {
		if p.PriceCents < 500 || p.PriceCents >= 200000 {
			t.Fatalf("pet %d price %d outside [500, 200000)", p.ID, p.PriceCents)
		}
	}
}
