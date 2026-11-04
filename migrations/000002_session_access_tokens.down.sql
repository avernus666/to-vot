DROP INDEX IF EXISTS sessions_access_active_idx;
ALTER TABLE sessions DROP CONSTRAINT IF EXISTS sessions_access_expiry_pair_check;
ALTER TABLE sessions DROP COLUMN IF EXISTS access_expires_at;
ALTER TABLE sessions DROP COLUMN IF EXISTS access_token_hash;
