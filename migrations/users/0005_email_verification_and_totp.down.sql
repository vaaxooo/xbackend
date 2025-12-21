DROP INDEX IF EXISTS idx_auth_verification_tokens_identity_type;
DROP TABLE IF EXISTS auth_verification_tokens;
ALTER TABLE auth_identities DROP COLUMN IF EXISTS email_confirmed_at;
ALTER TABLE auth_identities DROP COLUMN IF EXISTS totp_secret;
ALTER TABLE auth_identities DROP COLUMN IF EXISTS totp_confirmed_at;
