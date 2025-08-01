-- +goose Up
-- SQL in this section is executed when the migration is applied
ALTER TABLE carts ADD COLUMN session_id VARCHAR(255) DEFAULT '';
CREATE INDEX idx_carts_session_id ON carts(session_id);
