CREATE TABLE IF NOT EXISTS auth_audit_events (
  id UUID PRIMARY KEY,
  account_id UUID NULL,
  event TEXT NOT NULL,
  at TIMESTAMPTZ NOT NULL,
  meta JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX IF NOT EXISTS idx_auth_audit_events_account_at
  ON auth_audit_events(account_id, at DESC);

CREATE INDEX IF NOT EXISTS idx_auth_audit_events_at
  ON auth_audit_events(at DESC);
