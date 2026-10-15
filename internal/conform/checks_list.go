package conform

import (
	"fmt"
	"net/http"
)

// The literal numbers below are worked out from the seed rules in
// contract/petshop.md, not computed with the generator.
var listChecks = []check{
	{"seed.counts", func(e *env) error {
		p, err := e.list("?page_size=1")
		if err != nil {
			return err
		}
		if p.Total != 10000 {
			return fmt.Errorf("total pets %d, want 10000 (is the database freshly seeded?)", p.Total)
		}
		return nil
	}},
	{"list.default", func(e *env) error {
		p, err := e.list("")
		if err != nil {
			return err
		}
		if p.Page != 1 || p.PageSize != 20 || len(p.Items) != 20 || p.Total != 10000 {
			return fmt.Errorf("page %d size %d items %d total %d, want 1/20/20/10000", p.Page, p.PageSize, len(p.Items), p.Total)
		}
		for i, it := range p.Items {
			if it.ID != i+1 {
				return fmt.Errorf("item %d has id %d, want %d (ordered by id ascending)", i, it.ID, i+1)
			}
		}
		return nil
	}},
	{"list.filter.category", func(e *env) error {
		return filtered(e, "?category=1&page_size=100", 500, func(p pet) bool { return p.CategoryID == 1 })
	}},
	{"list.filter.status", func(e *env) error {
		return filtered(e, "?status=sold&page_size=100", 1500, func(p pet) bool { return p.Status == "sold" })
	}},
	{"list.filter.price", func(e *env) error {
		return filtered(e, "?min_price=10000&max_price=50000&page_size=100", 2006, func(p pet) bool {
			return p.PriceCents >= 10000 && p.PriceCents <= 50000
		})
	}},
	{"list.filter.combined", func(e *env) error {
		return filtered(e, "?category=3&status=available&min_price=100000&page_size=100", 250, func(p pet) bool {
			return p.CategoryID == 3 && p.Status == "available" && p.PriceCents >= 100000
		})
	}},
	{"list.paging", func(e *env) error {
		p, err := e.list("?page=2&page_size=100")
		if err != nil {
			return err
		}
		if p.Page != 2 || p.PageSize != 100 || len(p.Items) != 100 || p.Items[0].ID != 101 {
			return fmt.Errorf("page 2 of 100: page %d size %d items %d", p.Page, p.PageSize, len(p.Items))
		}
		return nil
	}},
	{"list.beyond_last", func(e *env) error {
		p, err := e.list("?page=101&page_size=100")
		if err != nil {
			return err
		}
		if len(p.Items) != 0 || p.Total != 10000 {
			return fmt.Errorf("past the last page: %d items, total %d, want 0 items and total 10000", len(p.Items), p.Total)
		}
		return nil
	}},
	{"list.invalid", func(e *env) error {
		for _, q := range []string{"?page=0", "?page_size=0", "?page_size=101", "?page_size=abc", "?status=bogus", "?category=x"} {
			if _, err := e.expect(http.StatusBadRequest, "GET", "/api/pets"+q, false, nil); err != nil {
				return err
			}
		}
		return nil
	}},
	{"pet.get", func(e *env) error {
		r, err := e.expect(http.StatusOK, "GET", "/api/pets/1", false, nil)
		if err != nil {
			return err
		}
		var p pet
		if err := decode(r, &p); err != nil {
			return err
		}
		if want := (pet{1, "Pet 00001", 8419, "available", 1}); p != want {
			return fmt.Errorf("pet 1 = %+v, want %+v", p, want)
		}
		return nil
	}},
	{"pet.get_missing", func(e *env) error {
		_, err := e.expect(http.StatusNotFound, "GET", "/api/pets/999999", false, nil)
		return err
	}},
}

// filtered requires the total and that every returned item matches the filter.
func filtered(e *env, query string, total int, match func(pet) bool) error {
	p, err := e.list(query)
	if err != nil {
		return err
	}
	if p.Total != total {
		return fmt.Errorf("%s: total %d, want %d", query, p.Total, total)
	}
	if len(p.Items) == 0 {
		return fmt.Errorf("%s: no items", query)
	}
	for _, it := range p.Items {
		if !match(it) {
			return fmt.Errorf("%s: pet %+v does not match the filter", query, it)
		}
	}
	return nil
}
