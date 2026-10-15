package shop

type Category struct {
	ID   int64  `tango:"pk" json:"id"`
	Name string `json:"name"`
}

type Pet struct {
	ID         int64  `tango:"pk" json:"id"`
	Name       string `json:"name"`
	PriceCents int64  `json:"price_cents"`
	Status     string `json:"status" tango:"index"`
	CategoryID int64  `tango:"fk=Category" json:"category_id"`
}

type Customer struct {
	ID    int64  `tango:"pk" json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Order struct {
	ID         int64 `tango:"pk" json:"id"`
	PetID      int64 `tango:"fk=Pet" json:"pet_id"`
	CustomerID int64 `tango:"fk=Customer" json:"customer_id"`
}
