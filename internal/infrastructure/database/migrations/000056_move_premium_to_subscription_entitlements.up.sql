CREATE TABLE IF NOT EXISTS subscription_entitlements (
  id SERIAL PRIMARY KEY,
  user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  provider VARCHAR(32) NOT NULL,
  product_id VARCHAR(128) NOT NULL,
  plan VARCHAR(32) NOT NULL DEFAULT 'monthly',
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  transaction_id VARCHAR(255),
  original_transaction_id VARCHAR(255),
  purchase_token_hash VARCHAR(128),
  expires_at TIMESTAMPTZ,
  environment VARCHAR(32) NOT NULL DEFAULT 'production',
  raw_payload JSONB,
  last_verified_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT subscription_entitlements_provider_check
    CHECK (provider IN ('apple', 'google', 'admin')),
  CONSTRAINT subscription_entitlements_status_check
    CHECK (status IN ('active', 'expired', 'cancelled', 'revoked', 'grace_period')),
  CONSTRAINT subscription_entitlements_plan_check
    CHECK (plan IN ('monthly'))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_subscription_entitlements_user_product
  ON subscription_entitlements(user_id, provider, product_id);

CREATE UNIQUE INDEX IF NOT EXISTS idx_subscription_entitlements_original_tx
  ON subscription_entitlements(provider, original_transaction_id)
  WHERE original_transaction_id IS NOT NULL AND original_transaction_id <> '';

CREATE UNIQUE INDEX IF NOT EXISTS idx_subscription_entitlements_purchase_token_hash
  ON subscription_entitlements(provider, purchase_token_hash)
  WHERE purchase_token_hash IS NOT NULL AND purchase_token_hash <> '';

CREATE INDEX IF NOT EXISTS idx_subscription_entitlements_active_user
  ON subscription_entitlements(user_id, status, expires_at);

INSERT INTO subscription_entitlements (
  user_id,
  provider,
  product_id,
  plan,
  status,
  expires_at,
  environment,
  last_verified_at,
  raw_payload
)
SELECT
  id,
  'admin',
  'notly_pro_monthly',
  'monthly',
  'active',
  premium_expires_at,
  'legacy',
  NOW(),
  jsonb_build_object('migrated_from', 'users.is_premium')
FROM users
WHERE is_premium = true
ON CONFLICT DO NOTHING;

DROP INDEX IF EXISTS idx_users_is_premium;

ALTER TABLE users
  DROP COLUMN IF EXISTS premium_expires_at,
  DROP COLUMN IF EXISTS premium_plan,
  DROP COLUMN IF EXISTS is_premium;
