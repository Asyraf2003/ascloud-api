package auth

const qInsertSession = `
	INSERT INTO auth_sessions (
		user_id, project_id,
		refresh_token_hash,
		device_id, user_agent_hash, ip_prefix,
		expires_at, meta
	) VALUES (
		$1, $2,
		$3,
		$4, $5, $6,
		$7, $8
	)
	RETURNING id::text, created_at
`

const qSelectSessionByID = `
	SELECT
		id::text,
		user_id::text,
		COALESCE(project_id::text, ''),
		refresh_token_hash,
		device_id,
		user_agent_hash,
		COALESCE(ip_prefix, ''),
		created_at,
		expires_at,
		revoked_at,
		meta
	FROM auth_sessions
	WHERE id = $1
`

const qSelectSessionByRefreshHash = `
	SELECT
		id::text,
		user_id::text,
		COALESCE(project_id::text, ''),
		refresh_token_hash,
		device_id,
		user_agent_hash,
		COALESCE(ip_prefix, ''),
		created_at,
		expires_at,
		revoked_at,
		meta
	FROM auth_sessions
	WHERE refresh_token_hash = $1
`

const qRotateRefreshHash = `
	UPDATE auth_sessions
	SET refresh_token_hash = $1,
	    expires_at = $2
	WHERE id = $3
	  AND refresh_token_hash = $4
	  AND revoked_at IS NULL
	  AND expires_at > now()
`

const qRevokeSession = `
	UPDATE auth_sessions
	SET revoked_at = $2
	WHERE id = $1
	  AND revoked_at IS NULL
`
