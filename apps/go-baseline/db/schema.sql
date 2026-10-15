CREATE TABLE IF NOT EXISTS categories (
    id   serial PRIMARY KEY,
    name text NOT NULL
);

CREATE TABLE IF NOT EXISTS pets (
    id          serial PRIMARY KEY,
    name        text NOT NULL,
    price_cents integer NOT NULL CHECK (price_cents >= 0),
    status      text NOT NULL CHECK (status IN ('available', 'pending', 'sold')),
    category_id integer NOT NULL REFERENCES categories (id)
);
CREATE INDEX IF NOT EXISTS pets_category_id ON pets (category_id);
CREATE INDEX IF NOT EXISTS pets_status ON pets (status);

CREATE TABLE IF NOT EXISTS customers (
    id    serial PRIMARY KEY,
    name  text NOT NULL,
    email text NOT NULL
);

CREATE TABLE IF NOT EXISTS orders (
    id          serial PRIMARY KEY,
    pet_id      integer NOT NULL REFERENCES pets (id),
    customer_id integer NOT NULL REFERENCES customers (id)
);

CREATE TABLE IF NOT EXISTS staff (
    username      text PRIMARY KEY,
    password_hash text NOT NULL
);
