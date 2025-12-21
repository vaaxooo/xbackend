ALTER TABLE auth_identities
    ADD COLUMN IF NOT EXISTS email_confirmed_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS totp_secret TEXT NULL,
    ADD COLUMN IF NOT EXISTS totp_confirmed_at TIMESTAMPTZ NULL;

CREATE TABLE IF NOT EXISTS auth_verification_tokens (
    id UUID PRIMARY KEY,
    identity_id UUID NOT NULL REFERENCES auth_identities(id) ON DELETE CASCADE,
    token_type TEXT NOT NULL,
    token_code TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_auth_verification_tokens_identity_type ON auth_verification_tokens(identity_id, token_type);
