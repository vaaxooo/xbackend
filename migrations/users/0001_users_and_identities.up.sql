CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    first_name TEXT NULL,
    last_name TEXT NULL,
    middle_name TEXT NULL,
    display_name TEXT NULL,
    avatar_url TEXT NULL,
    profile_customized BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS auth_identities (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    provider TEXT NOT NULL,              -- "email", "telegram", "google", "discord", "phone"
    provider_user_id TEXT NOT NULL,      -- email / telegram_id / google_sub / discord_id / phone
    secret_hash TEXT NULL,               -- password hash for email; NULL for oauth providers
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (provider, provider_user_id)
);

CREATE INDEX IF NOT EXISTS idx_auth_identities_user_id ON auth_identities(user_id);
