package domain

const (
	ErrOIDCStateInvalid   = "AUTH_OIDC_STATE_INVALID"
	ErrOIDCExchangeFailed = "AUTH_OIDC_EXCHANGE_FAILED"
	ErrOIDCIDTokenInvalid = "AUTH_OIDC_IDTOKEN_INVALID" // #nosec G101 -- error code identifier, not a credential

	ErrEmailNotVerified = "AUTH_EMAIL_NOT_VERIFIED"

	ErrUnauthorized   = "AUTH_UNAUTHORIZED"
	ErrRefreshMissing = "AUTH_REFRESH_MISSING"
	ErrRefreshReused  = "AUTH_REFRESH_REUSED"

	ErrCSRFInvalid    = "AUTH_CSRF_FAILED"
	ErrStepUpRequired = "AUTH_STEP_UP_REQUIRED"
	ErrForbidden      = "AUTH_FORBIDDEN"
	ErrBadRequest     = "AUTH_BAD_REQUEST"
	ErrInternal       = "AUTH_INTERNAL"
)
