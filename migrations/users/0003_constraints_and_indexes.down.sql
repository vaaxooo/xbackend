ALTER TABLE auth_identities
    DROP CONSTRAINT IF EXISTS uq_auth_identities_user_provider;

DROP INDEX IF EXISTS idx_auth_refresh_tokens_user_id;
DROP INDEX IF EXISTS idx_auth_refresh_tokens_expires_at;
DROP INDEX IF EXISTS idx_auth_refresh_tokens_revoked_at;
