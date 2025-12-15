ALTER TABLE auth_sessions
  DROP CONSTRAINT IF EXISTS auth_sessions_user_id_fkey;

ALTER TABLE auth_sessions
  ADD CONSTRAINT auth_sessions_user_id_fkey
  FOREIGN KEY (user_id) REFERENCES accounts(id)
  ON DELETE CASCADE;
