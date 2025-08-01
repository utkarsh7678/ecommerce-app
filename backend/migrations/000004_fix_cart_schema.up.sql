-- Disable foreign key checks temporarily
PRAGMA foreign_keys = OFF;

-- Create new tables with proper schema
CREATE TABLE IF NOT EXISTS new_carts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS new_cart_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    cart_id INTEGER NOT NULL,
    item_id INTEGER NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (cart_id) REFERENCES carts(id) ON DELETE CASCADE,
    FOREIGN KEY (item_id) REFERENCES items(id) ON DELETE CASCADE,
    UNIQUE(cart_id, item_id)
);

-- Copy data from old tables to new ones, ensuring referential integrity
-- Only copy carts that have a valid user
INSERT INTO new_carts (id, user_id, status, created_at, updated_at)
SELECT c.id, c.user_id, 
       COALESCE(c.status, 'active') as status,
       COALESCE(c.created_at, CURRENT_TIMESTAMP) as created_at,
       COALESCE(c.updated_at, CURRENT_TIMESTAMP) as updated_at
FROM carts c
WHERE EXISTS (SELECT 1 FROM users u WHERE u.id = c.user_id);

-- Only copy cart items that have valid cart and item references
INSERT INTO new_cart_items (id, cart_id, item_id, quantity, created_at, updated_at)
SELECT ci.id, ci.cart_id, ci.item_id, 
       COALESCE(ci.quantity, 1) as quantity,
       COALESCE(ci.created_at, CURRENT_TIMESTAMP) as created_at,
       COALESCE(ci.updated_at, CURRENT_TIMESTAMP) as updated_at
FROM cart_items ci
WHERE EXISTS (SELECT 1 FROM new_carts c WHERE c.id = ci.cart_id)
  AND EXISTS (SELECT 1 FROM items i WHERE i.id = ci.item_id);

-- Drop old tables
DROP TABLE IF EXISTS cart_items;
DROP TABLE IF EXISTS carts;

-- Rename new tables to original names
ALTER TABLE new_carts RENAME TO carts;
ALTER TABLE new_cart_items RENAME TO cart_items;

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_carts_user_id ON carts(user_id);
CREATE INDEX IF NOT EXISTS idx_cart_items_cart_id ON cart_items(cart_id);
CREATE INDEX IF NOT EXISTS idx_cart_items_item_id ON cart_items(item_id);

-- Enable foreign key checks
PRAGMA foreign_keys = ON;
