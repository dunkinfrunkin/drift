-- Pre-existing tables that drift does not manage.
-- These simulate a database that already has data before drift is introduced.

CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    price NUMERIC(10, 2) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(id),
    quantity INTEGER NOT NULL DEFAULT 1,
    ordered_at TIMESTAMP NOT NULL DEFAULT NOW()
);

INSERT INTO products (name, price) VALUES
    ('Widget', 9.99),
    ('Gadget', 24.99),
    ('Doohickey', 4.50);

INSERT INTO orders (product_id, quantity) VALUES
    (1, 10),
    (2, 3),
    (3, 7);
