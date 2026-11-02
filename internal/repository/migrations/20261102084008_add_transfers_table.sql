-- +goose Up
CREATE TABLE transfers(
  id SERIAL PRIMARY KEY,
  from_user_id INTEGER NOT NULL REFERENCES users(id),
  to_user_id INTEGER NOT NULL REFERENCES users(id),
  amount INTEGER NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  CONSTRAINT amount_nonnegative CHECK(amount >= 0),
  CONSTRAINT amount_different_users CHECK(from_user_id <> to_user_id)
);

-- +goose Down
DROP TABLE transfers;
