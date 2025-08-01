-- This is a simplified rollback that preserves data but may not restore all constraints
PRAGMA foreign_keys=off;

-- Recreate original cart_items structure
CREATE TABLE IF NOT EXISTS cart_items_backup (
  cart_id INTEGER NOT NULL,
  item_id INTEGER NOT NULL,
  quantity INTEGER DEFAULT 1,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  PRIMARY KEY (cart_id, item_id)
);

-- Copy data back to original structure
INSERT INTO cart_items_backup (cart_id, item_id, quantity, created_at, updated_at)
SELECT cart_id, item_id, quantity, created_at, updated_at
FROM cart_items;

-- Drop and recreate the table
DROP TABLE cart_items;
ALTER TABLE cart_items_backup RENAME TO cart_items;

-- Recreate original carts structure
CREATE TABLE IF NOT EXISTS carts_backup (
  id INTEGER PRIMARY KEY,
  user_id INTEGER NOT NULL,
  status TEXT,
  created_at TIMESTAMP,
  updated_at TIMESTAMP
);

-- Copy data back to original structure
INSERT INTO carts_backup (id, user_id, status, created_at, updated_at)
SELECT id, user_id, status, created_at, updated_at
FROM carts;

-- Drop and recreate the table
DROP TABLE carts;
ALTER TABLE carts_backup RENAME TO carts;

-- Recreate indexes
CREATE INDEX IF NOT EXISTS idx_cart_items_cart_id ON cart_items(cart_id);

PRAGMA foreign_keys=on;
