-- Drop and recreate carts table with proper schema
CREATE TABLE IF NOT EXISTS carts_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Copy data from old carts table if it exists
INSERT INTO carts_new (id, user_id, status, created_at, updated_at)
SELECT id, user_id, COALESCE(status, 'active'), 
       COALESCE(created_at, CURRENT_TIMESTAMP), 
       COALESCE(updated_at, CURRENT_TIMESTAMP)
FROM carts WHERE 1=1;

-- Drop old carts table and rename new one
DROP TABLE IF EXISTS carts_old;
ALTER TABLE carts RENAME TO carts_old;
ALTER TABLE carts_new RENAME TO carts;

-- Create cart_items table with proper schema
CREATE TABLE IF NOT EXISTS cart_items_new (
    cart_id INTEGER NOT NULL,
    item_id INTEGER NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (cart_id, item_id),
    FOREIGN KEY (cart_id) REFERENCES carts(id) ON DELETE CASCADE,
    FOREIGN KEY (item_id) REFERENCES items(id) ON DELETE CASCADE
);

-- Copy data from old cart_items table if it exists
INSERT INTO cart_items_new (cart_id, item_id, quantity, created_at, updated_at)
SELECT cart_id, item_id, COALESCE(quantity, 1), 
       COALESCE(created_at, CURRENT_TIMESTAMP), 
       COALESCE(updated_at, CURRENT_TIMESTAMP)
FROM cart_items WHERE 1=1;

-- Drop old cart_items table and rename new one
DROP TABLE IF EXISTS cart_items_old;
ALTER TABLE cart_items RENAME TO cart_items_old;
ALTER TABLE cart_items_new RENAME TO cart_items;

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_cart_items_cart_id ON cart_items(cart_id);
CREATE INDEX IF NOT EXISTS idx_cart_items_item_id ON cart_items(item_id);
CREATE INDEX IF NOT EXISTS idx_carts_user_id ON carts(user_id);

-- Clean up old tables after successful migration
DROP TABLE IF EXISTS carts_old;
DROP TABLE IF EXISTS cart_items_old;
