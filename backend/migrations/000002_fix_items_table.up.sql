-- Add any missing columns to items table if they don't exist
ALTER TABLE items 
    ADD COLUMN IF NOT EXISTS id SERIAL PRIMARY KEY,
    ADD COLUMN IF NOT EXISTS name VARCHAR(255) NOT NULL,
    ADD COLUMN IF NOT EXISTS status VARCHAR(50) DEFAULT 'available',
    ADD COLUMN IF NOT EXISTS price DECIMAL(10,2) NOT NULL,
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;

-- Add any missing indexes
CREATE INDEX IF NOT EXISTS idx_items_id ON items(id);

-- Ensure the cart_items table has the correct structure
ALTER TABLE cart_items
    ADD COLUMN IF NOT EXISTS id SERIAL PRIMARY KEY,
    ADD COLUMN IF NOT EXISTS cart_id INTEGER NOT NULL,
    ADD COLUMN IF NOT EXISTS item_id INTEGER NOT NULL,
    ADD COLUMN IF NOT EXISTS quantity INTEGER NOT NULL DEFAULT 1,
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;

-- Add foreign key constraints if they don't exist
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_cart_items_cart') THEN
        ALTER TABLE cart_items ADD CONSTRAINT fk_cart_items_cart 
            FOREIGN KEY (cart_id) REFERENCES carts(id) ON DELETE CASCADE;
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_cart_items_item') THEN
        ALTER TABLE cart_items ADD CONSTRAINT fk_cart_items_item 
            FOREIGN KEY (item_id) REFERENCES items(id) ON DELETE CASCADE;
    END IF;
END
$$;
