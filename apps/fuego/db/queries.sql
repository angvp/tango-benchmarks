-- name: CountPets :one
SELECT count(*) FROM pets
WHERE (sqlc.narg('category')::int IS NULL OR category_id = sqlc.narg('category'))
  AND (sqlc.narg('status')::text IS NULL OR status = sqlc.narg('status'))
  AND (sqlc.narg('min_price')::int IS NULL OR price_cents >= sqlc.narg('min_price'))
  AND (sqlc.narg('max_price')::int IS NULL OR price_cents <= sqlc.narg('max_price'));

-- name: ListPets :many
SELECT id, name, price_cents, status, category_id FROM pets
WHERE (sqlc.narg('category')::int IS NULL OR category_id = sqlc.narg('category'))
  AND (sqlc.narg('status')::text IS NULL OR status = sqlc.narg('status'))
  AND (sqlc.narg('min_price')::int IS NULL OR price_cents >= sqlc.narg('min_price'))
  AND (sqlc.narg('max_price')::int IS NULL OR price_cents <= sqlc.narg('max_price'))
ORDER BY id
LIMIT sqlc.arg('page_size')::int OFFSET sqlc.arg('skip')::int;

-- name: GetPet :one
SELECT id, name, price_cents, status, category_id FROM pets WHERE id = $1;

-- name: CreatePet :one
INSERT INTO pets (name, price_cents, status, category_id) VALUES ($1, $2, $3, $4)
RETURNING id, name, price_cents, status, category_id;

-- name: UpdatePet :one
UPDATE pets SET name = $2, price_cents = $3, status = $4, category_id = $5 WHERE id = $1
RETURNING id, name, price_cents, status, category_id;

-- name: CategoryExists :one
SELECT EXISTS (SELECT 1 FROM categories WHERE id = $1);

-- name: PetExists :one
SELECT EXISTS (SELECT 1 FROM pets WHERE id = $1);

-- name: CustomerExists :one
SELECT EXISTS (SELECT 1 FROM customers WHERE id = $1);

-- name: CreateOrder :one
INSERT INTO orders (pet_id, customer_id) VALUES ($1, $2) RETURNING id, pet_id, customer_id;

-- name: GetOrder :one
SELECT id, pet_id, customer_id FROM orders WHERE id = $1;

-- name: GetStaffHash :one
SELECT password_hash FROM staff WHERE username = $1;

-- name: UpsertStaff :exec
INSERT INTO staff (username, password_hash) VALUES ($1, $2)
ON CONFLICT (username) DO UPDATE SET password_hash = EXCLUDED.password_hash;
