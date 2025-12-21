DROP TABLE IF EXISTS auth_challenges;
ALTER TABLE users
    DROP COLUMN IF EXISTS suspended,
    DROP COLUMN IF EXISTS suspension_reason,
    DROP COLUMN IF EXISTS blocked_until;
