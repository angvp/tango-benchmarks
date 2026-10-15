// The uAdmin pet shop: uAdmin's models and admin, with the contract's JSON
// endpoints written as ordinary handlers on the same mux.
package main

import (
	"os"

	"github.com/uadmin/uadmin"
)

const staffUsername = "bench-staff"

func main() {
	uadmin.Register(Category{}, Pet{}, Customer{}, Order{})
	ensureStaff()
	registerAPI()
	uadmin.StartServer()
}

// ensureStaff creates the benchmark staff account on first start.
func ensureStaff() {
	if uadmin.Count(&uadmin.User{}, "username = ?", staffUsername) > 0 {
		return
	}
	password := os.Getenv("BENCH_STAFF_PASSWORD")
	if password == "" {
		password = "bench-staff-password"
	}
	staff := uadmin.User{Username: staffUsername, FirstName: "Bench", LastName: "Staff", Password: password, Active: true, Admin: true}
	staff.Save()
}
