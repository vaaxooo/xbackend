CREATE INDEX IF NOT EXISTS idx_auth_refresh_tokens_fingerprint
    ON auth_refresh_tokens(user_id, user_agent, ip);
