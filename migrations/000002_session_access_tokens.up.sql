ALTER TABLE sessions
    ADD COLUMN access_token_hash BYTEA,
    ADD COLUMN access_expires_at TIMESTAMPTZ;

CREATE INDEX sessions_access_active_idx
    ON sessions(access_token_hash, access_expires_at)
    WHERE revoked_at IS NULL AND access_token_hash IS NOT NULL;

ALTER TABLE sessions
    ADD CONSTRAINT sessions_access_expiry_pair_check
    CHECK (
        (access_token_hash IS NULL AND access_expires_at IS NULL)
        OR
        (access_token_hash IS NOT NULL AND access_expires_at IS NOT NULL)
    );
