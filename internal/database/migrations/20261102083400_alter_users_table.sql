-- +goose Up
ALTER TABLE users
ADD COLUMN id SERIAL PRIMARY KEY,
ADD CONSTRAINT users_balance_nonnegative CHECK(balance >= 0);

-- +goose Down
ALTER TABLE users
DROP COLUMN id SERIAL PRIMARY KEY,
DROP CONSTRAINT users_balance_nonnegative;
