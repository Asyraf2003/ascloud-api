CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Ensure accounts exists (harusnya sudah dari 0006, tapi aman)
CREATE TABLE IF NOT EXISTS accounts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email TEXT NOT NULL UNIQUE,
  meta JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Reconcile auth_sessions (kalau sudah ada versi lama)
ALTER TABLE IF EXISTS auth_sessions
  ADD COLUMN IF NOT EXISTS project_id UUID NULL,
  ADD COLUMN IF NOT EXISTS refresh_token_hash TEXT,
  ADD COLUMN IF NOT EXISTS device_id TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS user_agent_hash TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS ip_prefix TEXT NULL,
  ADD COLUMN IF NOT EXISTS expires_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS revoked_at TIMESTAMPTZ NULL,
  ADD COLUMN IF NOT EXISTS meta JSONB NOT NULL DEFAULT '{}'::jsonb;

CREATE UNIQUE INDEX IF NOT EXISTS ux_auth_sessions_refresh_hash
  ON auth_sessions(refresh_token_hash);

CREATE UNIQUE INDEX IF NOT EXISTS ux_auth_sessions_one_active_per_user
  ON auth_sessions(user_id)
  WHERE revoked_at IS NULL;

-- auth_refresh_used
CREATE TABLE IF NOT EXISTS auth_refresh_used (
  hash TEXT PRIMARY KEY,
  session_id UUID NOT NULL,
  used_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_auth_refresh_used_session
  ON auth_refresh_used(session_id);

-- FK auth_sessions.user_id -> accounts(id)
ALTER TABLE auth_sessions
  DROP CONSTRAINT IF EXISTS auth_sessions_user_id_fkey;

ALTER TABLE auth_sessions
  ADD CONSTRAINT auth_sessions_user_id_fkey
  FOREIGN KEY (user_id) REFERENCES accounts(id)
  ON DELETE CASCADE;

-- FK auth_identities.account_id -> accounts(id)
ALTER TABLE auth_identities
  DROP CONSTRAINT IF EXISTS auth_identities_account_id_fkey;

ALTER TABLE auth_identities
  ADD CONSTRAINT auth_identities_account_id_fkey
  FOREIGN KEY (account_id) REFERENCES accounts(id)
  ON DELETE CASCADE;

-- FK auth_refresh_used.session_id -> auth_sessions(id)
ALTER TABLE auth_refresh_used
  DROP CONSTRAINT IF EXISTS auth_refresh_used_session_id_fkey;

ALTER TABLE auth_refresh_used
  ADD CONSTRAINT auth_refresh_used_session_id_fkey
  FOREIGN KEY (session_id) REFERENCES auth_sessions(id)
  ON DELETE CASCADE;
