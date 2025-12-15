package jwt

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type Verifier struct {
	issuer string
	aud    string
	secret []byte
}

func NewHMACVerifier(issuer, aud, secret string) (*Verifier, error) {
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("jwt secret empty")
	}
	return &Verifier{issuer: issuer, aud: aud, secret: []byte(secret)}, nil
}

func (v *Verifier) Verify(token string) (Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return Claims{}, errInvalidToken
	}

	var h Header
	if err := b64urlDecodeJSON(parts[0], &h); err != nil || h.Alg != "HS256" {
		return Claims{}, errInvalidToken
	}

	input := parts[0] + "." + parts[1]
	if !verifyHS256(v.secret, input, parts[2]) {
		return Claims{}, errInvalidToken
	}

	var p Payload
	if err := b64urlDecodeJSON(parts[1], &p); err != nil {
		return Claims{}, errInvalidToken
	}
	if p.Iss != v.issuer || p.Aud != v.aud || p.Sub == "" || p.Sid == "" {
		return Claims{}, errInvalidToken
	}
	if time.Now().Unix() >= p.EXP {
		return Claims{}, fmt.Errorf("expired: %w", errInvalidToken)
	}

	return Claims{
		AccountID: p.Sub, SessionID: p.Sid, TrustLevel: p.AAL,
		JWTID: p.JTI, Issuer: p.Iss, Aud: p.Aud, IssuedAt: p.IAT, ExpiresAt: p.EXP,
	}, nil
}
