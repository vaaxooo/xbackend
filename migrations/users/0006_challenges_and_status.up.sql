ALTER TABLE users
    ADD COLUMN IF NOT EXISTS suspended BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS suspension_reason TEXT NULL,
    ADD COLUMN IF NOT EXISTS blocked_until TIMESTAMPTZ NULL;

CREATE TABLE IF NOT EXISTS auth_challenges (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    challenge_type TEXT NOT NULL,
    required_steps TEXT[] NOT NULL,
    completed_steps TEXT[] NOT NULL,
    status TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    session_fingerprint TEXT NULL,
    attempts_left INT NOT NULL DEFAULT 0,
    lock_until TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_auth_challenges_user ON auth_challenges(user_id);
