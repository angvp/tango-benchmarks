package main

import "github.com/uadmin/uadmin"

type Category struct {
	uadmin.Model
	Name string
}

func (c Category) String() string { return c.Name }

type Pet struct {
	uadmin.Model
	Name       string
	PriceCents int
	Status     string `uadmin:"filter"`
	Category   Category
	CategoryID uint
}

func (p Pet) String() string { return p.Name }

type Customer struct {
	uadmin.Model
	Name  string
	Email string
}

func (c Customer) String() string { return c.Name }

type Order struct {
	uadmin.Model
	Pet        Pet
	PetID      uint
	Customer   Customer
	CustomerID uint
}
