CREATE TABLE users (
    user_id SERIAL PRIMARY KEY,
    name VARCHAR(100)
);

CREATE TABLE balances (
    PRIMARY KEY (user_id),
    user_id INT REFERENCES users(user_id) ON DELETE CASCADE,
    balance NUMERIC(10, 2) DEFAULT 0 CHECK (balance >= 0)
);

CREATE TABLE services (
    service_id SERIAL PRIMARY KEY,
    service_name VARCHAR(100) NOT NULL
);

CREATE TABLE transactions (
    transaction_id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(user_id),
    service_id INT REFERENCES services(service_id),
    amount NUMERIC(10, 2) NOT NULL CHECK (amount >= 0),
    transaction_type VARCHAR(10) NOT NULL, 
    transaction_date TIMESTAMP DEFAULT NOW(),
    comment TEXT
);

CREATE TABLE reserved_funds (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(user_id),
    service_id INTEGER REFERENCES services(service_id),
    order_id INTEGER,
    amount DECIMAL(20, 2) NOT NULL CHECK (amount >= 0), 
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_transactions_date ON transactions (transaction_date);

