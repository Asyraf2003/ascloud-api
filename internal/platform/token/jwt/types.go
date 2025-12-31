package jwt

import "example.com/your-api/internal/shared/authn"

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

type Claims = authn.Claims
