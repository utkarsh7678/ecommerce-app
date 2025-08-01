-- This migration is not easily reversible as it restructures the tables
-- The best we can do is preserve the data but lose some constraints

-- Create backup of cart_items with old schema
CREATE TABLE IF NOT EXISTS cart_items_backup AS SELECT * FROM cart_items;

-- Create backup of carts with old schema
CREATE TABLE IF NOT EXISTS carts_backup AS SELECT * FROM carts;

-- Note: To fully restore, you would need to manually recreate the original schema
-- and copy data back from the backup tables, then drop the backup tables

-- The actual rollback will just log that manual intervention is needed
SELECT 'WARNING: This migration requires manual rollback. Data has been backed up to cart_items_backup and carts_backup tables.' AS message;
