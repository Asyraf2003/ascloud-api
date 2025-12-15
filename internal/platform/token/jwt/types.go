package jwt

type Header struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
	Kid string `json:"kid,omitempty"`
}

type Payload struct {
	Iss string `json:"iss"`
	Aud string `json:"aud"`
	Sub string `json:"sub"`
	Sid string `json:"sid"`
	AAL string `json:"aal"`

	JTI string `json:"jti"`
	IAT int64  `json:"iat"`
	EXP int64  `json:"exp"`
}

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
