-- auth_identities: forbid linking same provider twice to the same user
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'uq_auth_identities_user_provider'
    ) THEN
        ALTER TABLE auth_identities
            ADD CONSTRAINT uq_auth_identities_user_provider UNIQUE (user_id, provider);
    END IF;
END $$;

-- refresh tokens: make lookups fast and enable cleanup
CREATE INDEX IF NOT EXISTS idx_auth_refresh_tokens_user_id ON auth_refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_auth_refresh_tokens_expires_at ON auth_refresh_tokens(expires_at);
CREATE INDEX IF NOT EXISTS idx_auth_refresh_tokens_revoked_at ON auth_refresh_tokens(revoked_at);
