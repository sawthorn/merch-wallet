-- +goose Up
CREATE TABLE inventory(
  user_id INTEGER NOT NULL REFERENCES users(id),
  item_id INTEGER NOT NULL REFERENCES items(id),
  quantity INTEGER NOT NULL,

  PRIMARY KEY(user_id, item_id),
  CONSTRAINT inventory_quantity_nonnegative CHECK(quantity >= 0)
);

-- +goose Down
DROP TABLE inventory;
