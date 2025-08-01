-- This migration cannot be cleanly rolled back due to potential data loss
-- from removing invalid references. This is a safety measure to prevent
-- data corruption.

-- Create backup tables before any changes
CREATE TABLE IF NOT EXISTS backup_carts AS SELECT * FROM carts;
CREATE TABLE IF NOT EXISTS backup_cart_items AS SELECT * FROM cart_items;

-- Note: The original tables will be recreated by previous migrations
-- when rolling back to an earlier state.
