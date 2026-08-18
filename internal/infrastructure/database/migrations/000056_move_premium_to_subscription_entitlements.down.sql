ALTER TABLE users
  ADD COLUMN IF NOT EXISTS is_premium BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS premium_plan VARCHAR(32) NOT NULL DEFAULT 'free',
  ADD COLUMN IF NOT EXISTS premium_expires_at TIMESTAMPTZ;

UPDATE users u
SET
  is_premium = true,
  premium_plan = COALESCE(e.plan, 'monthly'),
  premium_expires_at = e.expires_at
FROM (
  SELECT DISTINCT ON (user_id)
    user_id,
    plan,
    expires_at
  FROM subscription_entitlements
  WHERE status IN ('active', 'grace_period')
    AND (expires_at IS NULL OR expires_at > NOW())
  ORDER BY user_id, expires_at DESC NULLS FIRST, updated_at DESC
) e
WHERE u.id = e.user_id;

CREATE INDEX IF NOT EXISTS idx_users_is_premium ON users(is_premium);

DROP TABLE IF EXISTS subscription_entitlements;
