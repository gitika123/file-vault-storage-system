BEGIN;
CREATE INDEX IF NOT EXISTS users_display_name_trgm_idx ON users USING GIN (lower(display_name) gin_trgm_ops);
COMMIT;
