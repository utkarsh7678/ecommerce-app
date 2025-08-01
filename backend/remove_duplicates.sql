-- Create a temporary table with unique items
CREATE TEMPORARY TABLE temp_items AS
SELECT MIN(id) as id, name, price, status, MIN(created_at) as created_at, MAX(updated_at) as updated_at
FROM items
GROUP BY LOWER(name), price, status;

-- Clear the original table
DELETE FROM items;

-- Copy back the unique items
INSERT INTO items (id, name, price, status, created_at, updated_at)
SELECT id, name, price, status, created_at, updated_at FROM temp_items;

-- Drop the temporary table
DROP TABLE temp_items;

-- Reset auto-increment counter
UPDATE sqlite_sequence SET seq = (SELECT MAX(id) FROM items) WHERE name = 'items';

-- Verify the fix
SELECT '=== Items after removing duplicates ===' as message;
SELECT id, name, price, status FROM items ORDER BY id;
