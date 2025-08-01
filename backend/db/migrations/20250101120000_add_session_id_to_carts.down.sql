-- +goose Down
-- SQL in this section is executed when the migration is rolled back
DROP INDEX IF EXISTS idx_carts_session_id;
ALTER TABLE carts DROP COLUMN session_id;
