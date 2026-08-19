ALTER TABLE subscription_entitlements
  DROP CONSTRAINT IF EXISTS subscription_entitlements_provider_check,
  DROP CONSTRAINT IF EXISTS subscription_entitlements_plan_check,
  DROP CONSTRAINT IF EXISTS subscription_entitlements_status_check;

ALTER TABLE subscription_entitlements
  ADD CONSTRAINT subscription_entitlements_provider_check
    CHECK (provider IN ('apple', 'google', 'admin')),
  ADD CONSTRAINT subscription_entitlements_plan_check
    CHECK (plan IN ('monthly')),
  ADD CONSTRAINT subscription_entitlements_status_check
    CHECK (status IN ('active', 'expired', 'cancelled', 'revoked', 'grace_period'));
