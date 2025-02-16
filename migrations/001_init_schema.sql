-- +goose Up
-- +goose StatementBegin
-- Создание таблицы пользователей
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR NOT NULL UNIQUE,
    password BYTEA NOT NULL,
    balance INT NOT NULL DEFAULT 1000
);
-- Создание таблицы товаров
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    name VARCHAR NOT NULL UNIQUE,
    price INT NOT NULL
);
-- Создание таблицы операций
CREATE TABLE operations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id INT NOT NULL REFERENCES users(id),
    amount INT NOT NULL,
    type VARCHAR NOT NULL CHECK (type IN ('purchase', 'transfer')),
    counterparty_id INT NULL REFERENCES users(id) DEFAULT NULL, -- ID отправителя/получателя при переводе
    product_id INT NULL REFERENCES products(id) DEFAULT NULL, -- ID товара, если покупка
    created_at TIMESTAMP DEFAULT NOW()
);
--Создание таблицы инвенторя
CREATE TABLE inventory (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id),
    product_id INT NOT NULL REFERENCES products(id),
    quantity INT NOT NULL CHECK (quantity > 0),
    UNIQUE (user_id, product_id)
);
--Создание индексов
CREATE INDEX idx_transactions_user_id ON operations(user_id);
CREATE INDEX idx_transactions_counterparty_id ON operations(counterparty_id);
CREATE INDEX idx_transactions_product_id ON operations(product_id);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_transactions_product_id;
DROP INDEX IF EXISTS idx_transactions_counterparty_id;
DROP INDEX IF EXISTS idx_transactions_user_id;
DROP TABLE IF EXISTS operations;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS inventory;
-- +goose StatementEnd