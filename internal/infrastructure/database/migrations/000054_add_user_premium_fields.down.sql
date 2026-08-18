DROP INDEX IF EXISTS idx_users_is_premium;

ALTER TABLE users
  DROP COLUMN IF EXISTS premium_expires_at,
  DROP COLUMN IF EXISTS premium_plan,
  DROP COLUMN IF EXISTS is_premium;
