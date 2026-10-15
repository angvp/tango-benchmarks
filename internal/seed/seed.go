// Package seed generates the data every pet shop starts from. It uses no
// randomness: each row is a pure function of its id, so the committed CSV files
// can be regenerated and compared byte for byte.
package seed

import "fmt"

const (
	CategoryCount = 20
	PetCount      = 10000
	CustomerCount = 1000
)

type Category struct {
	ID   int
	Name string
}

type Pet struct {
	ID         int
	Name       string
	PriceCents int
	Status     string
	CategoryID int
}

type Customer struct {
	ID    int
	Name  string
	Email string
}

func Categories() []Category {
	out := make([]Category, 0, CategoryCount)
	for id := 1; id <= CategoryCount; id++ {
		out = append(out, Category{ID: id, Name: fmt.Sprintf("Category %02d", id)})
	}
	return out
}

// Pets: category cycles 1..20; status follows id mod 20 (0-13 available,
// 14-16 pending, 17-19 sold); price is 500 plus a fixed stride modulo 199500.
func Pets() []Pet {
	out := make([]Pet, 0, PetCount)
	for id := 1; id <= PetCount; id++ {
		out = append(out, Pet{
			ID:         id,
			Name:       fmt.Sprintf("Pet %05d", id),
			PriceCents: 500 + (id*7919)%199500,
			Status:     statusFor(id % 20),
			CategoryID: (id-1)%CategoryCount + 1,
		})
	}
	return out
}

func statusFor(residue int) string {
	switch {
	case residue <= 13:
		return "available"
	case residue <= 16:
		return "pending"
	default:
		return "sold"
	}
}

func Customers() []Customer {
	out := make([]Customer, 0, CustomerCount)
	for id := 1; id <= CustomerCount; id++ {
		out = append(out, Customer{
			ID:    id,
			Name:  fmt.Sprintf("Customer %04d", id),
			Email: fmt.Sprintf("customer%04d@example.test", id),
		})
	}
	return out
}
