CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS auth_sessions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL,
  project_id UUID NULL,

  refresh_token_hash TEXT NOT NULL,
  device_id TEXT NOT NULL DEFAULT '',
  user_agent_hash TEXT NOT NULL DEFAULT '',
  ip_prefix TEXT NULL,

  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  expires_at TIMESTAMPTZ NOT NULL,
  revoked_at TIMESTAMPTZ NULL,

  meta JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_auth_sessions_refresh_hash
  ON auth_sessions(refresh_token_hash);

CREATE UNIQUE INDEX IF NOT EXISTS ux_auth_sessions_one_active_per_user
  ON auth_sessions(user_id)
  WHERE revoked_at IS NULL;

CREATE TABLE IF NOT EXISTS auth_refresh_used (
  hash TEXT PRIMARY KEY,
  session_id UUID NOT NULL,
  used_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_auth_refresh_used_session
  ON auth_refresh_used(session_id);
