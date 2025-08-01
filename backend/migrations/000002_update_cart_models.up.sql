-- Drop and recreate cart_items table with proper constraints
PRAGMA foreign_keys=off;

-- Drop foreign key constraints first
DROP TABLE IF EXISTS cart_items_backup;

-- Create new cart_items table with proper schema
CREATE TABLE IF NOT EXISTS cart_items_backup (
  cart_id INTEGER NOT NULL,
  item_id INTEGER NOT NULL,
  quantity INTEGER DEFAULT 1,
  created_at TIMESTAMP NOT NULL DEFAULT (datetime('now')),
  updated_at TIMESTAMP NOT NULL DEFAULT (datetime('now', 'localtime')),
  PRIMARY KEY (cart_id, item_id),
  FOREIGN KEY (cart_id) REFERENCES carts(id) ON DELETE CASCADE,
  FOREIGN KEY (item_id) REFERENCES items(id) ON DELETE CASCADE
);

-- Copy data from old table to new table
INSERT INTO cart_items_backup (cart_id, item_id, quantity, created_at, updated_at)
SELECT cart_id, item_id, quantity, COALESCE(created_at, datetime('now')), COALESCE(updated_at, datetime('now', 'localtime'))
FROM cart_items;

-- Drop old table and rename new one
DROP TABLE cart_items;
ALTER TABLE cart_items_backup RENAME TO cart_items;

-- Update carts table to ensure proper autoincrement
CREATE TABLE IF NOT EXISTS carts_backup (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL,
  status TEXT DEFAULT 'active',
  created_at TIMESTAMP NOT NULL DEFAULT (datetime('now')),
  updated_at TIMESTAMP NOT NULL DEFAULT (datetime('now', 'localtime')),
  FOREIGN KEY (user_id) REFERENCES users(id)
);

-- Copy data from old carts table
INSERT INTO carts_backup (id, user_id, status, created_at, updated_at)
SELECT id, user_id, COALESCE(status, 'active'), 
       COALESCE(created_at, datetime('now')), 
       COALESCE(updated_at, datetime('now', 'localtime'))
FROM carts;

-- Drop old table and rename new one
DROP TABLE carts;
ALTER TABLE carts_backup RENAME TO carts;

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_cart_items_cart_id ON cart_items(cart_id);
CREATE INDEX IF NOT EXISTS idx_cart_items_item_id ON cart_items(item_id);
CREATE INDEX IF NOT EXISTS idx_carts_user_id ON carts(user_id);

PRAGMA foreign_keys=on;
