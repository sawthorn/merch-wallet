-- +goose Up
BEGIN;
CREATE TABLE items(
  id SERIAL PRIMARY KEY,
  name TEXT UNIQUE NOT NULL,
  cost INTEGER NOT NULL,

  CONSTRAINT items_cost_nonnegative CHECK(cost >= 0)
);
INSERT INTO items(name, cost) VALUES
('t-shirt',     80),
('cup',         20),
('book',        50),
('pen',         10),
('powerbank',   200),
('hoody',       300),
('umbrella',    200),
('socks',       10),
('wallet',      50),
('pink-hoody',  500);
COMMIT;

-- +goose Down
DROP TABLE items;
