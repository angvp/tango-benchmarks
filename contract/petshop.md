# The pet shop contract

Every framework in this repository builds the same small application. This page is the whole contract: the conformance suite (`cmd/conform`) checks exactly what is written here, black box, over HTTP. A framework that cannot meet a part says so with a reason; it is recorded as N/A, never as a pass and never as zero.

## Data

| Model | Fields |
|---|---|
| `Category` | `id`, `name` |
| `Pet` | `id`, `name`, `price_cents` (integer), `status` (`available`, `pending` or `sold`), `category_id` (many-to-one to `Category`) |
| `Customer` | `id`, `name`, `email` |
| `Order` | `id`, `pet_id` (many-to-one to `Pet`), `customer_id` (many-to-one to `Customer`) |

Each order is for one pet. There are no uploads, email, payments or many-to-many relations.

### Seed

Every app starts from the files in `seed/data` (20 categories, 10,000 pets, 1,000 customers; no orders), loaded by `cmd/loadseed` into the tables the app created, using the app's own `seed-map.json` to map CSV columns to table and column names. The suite expects a freshly seeded database: run `bench.sh reset` before each run.

## HTTP API

All bodies are JSON (`Content-Type: application/json`). Errors are `{"error": "<message>"}`.

### `POST /api/login`

Body `{"username": "...", "password": "..."}`. `200` with `{"token": "<opaque string>"}`, or `401`. The staff account is `bench-staff` with the password in the `BENCH_STAFF_PASSWORD` environment variable (default `bench-staff-password`; a benchmark credential, not a secret). Every app creates it during its own setup.

Writes and order reads send `Authorization: Bearer <token>`. Without a valid token they answer `401`.

### Pets

- `GET /api/pets`: public. Query parameters, all optional: `category` (id), `status`, `min_price`, `max_price` (cents, inclusive), `page` (default 1), `page_size` (default 20, at most 100). Ordered by `id` ascending. `200` with `{"items": [Pet...], "page": n, "page_size": n, "total": n}`, where `total` counts all matches. A page past the last returns `items: []` with the real `total`. `page` below 1, `page_size` outside 1-100, a non-numeric number or an unknown `status` answer `400`.
- `GET /api/pets/{id}`: public. `200` with the pet, or `404`.
- `POST /api/pets`: authenticated. Body `{"name", "price_cents", "status", "category_id"}`. `201` with the created pet (including `id`).
- `PUT /api/pets/{id}`: authenticated. Replaces all four fields. `200` with the pet, or `404`.

Both writes answer `422` when `name` is empty, `price_cents` is negative, `status` is not one of the three values, or `category_id` names no category. A missing token answers `401` before anything else, and an unknown id answers `404` before validation.

A pet is `{"id", "name", "price_cents", "status", "category_id"}`.

### Orders

- `POST /api/orders`: authenticated. Body `{"pet_id", "customer_id"}`. `201` with `{"id", "pet_id", "customer_id"}`, or `422` when either does not exist.
- `GET /api/orders/{id}`: authenticated. `200` or `404`.

## Admin (optional)

An app that ships an administration interface describes it in `app.json`:

```json
{"admin": {"login_path": "/admin/login/", "list_path": "/admin/petshop/pet/",
           "username_field": "username", "password_field": "password"}}
```

or `{"admin": null, "admin_na_reason": "..."}`. The suite checks that the list page answers neither `200` with pet rows nor any pet data to an anonymous client, then signs in through the HTML form (hidden inputs such as CSRF tokens are copied as the browser would) and checks that `list_path` answers `200` HTML containing the name of pet 1 (`Pet 00001`).

## Configuration

An app reads `DATABASE_URL` (a PostgreSQL URL), `LISTEN_ADDR` (default `:8080`) and `BENCH_STAFF_PASSWORD`. It runs its own migrations at start or through a documented setup step, which counts in the effort track.
