ALTER TABLE transactions RENAME COLUMN price_in_rub TO price;
ALTER TABLE transactions ADD COLUMN currency VARCHAR(3) NOT NULL DEFAULT 'RUB';
ALTER TABLE transactions ALTER COLUMN currency DROP DEFAULT;
