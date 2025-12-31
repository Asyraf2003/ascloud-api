package authn

type Claims struct {
	AccountID  string
	SessionID  string
	TrustLevel string

	JWTID  string
	Issuer string
	Aud    string

	IssuedAt  int64
	ExpiresAt int64
}

type AccessTokenVerifier interface {
	Verify(token string) (Claims, error)
}
