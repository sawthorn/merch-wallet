-- +goose Up
CREATE TABLE users(
  username TEXT UNIQUE NOT NULL,
  password TEXT NOT NULL,
  balance INTEGER NOT NULL
);

-- +goose Down
DROP TABLE users;
