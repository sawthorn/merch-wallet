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
CREATE INDEX idx_transfers_from_user_created_at
  ON transfers(from_user_id, created_at DESC);
CREATE INDEX idx_transfers_to_user_created_at
  ON transfers(to_user_id, created_at DESC);

-- +goose Down
DROP TABLE transfers;
