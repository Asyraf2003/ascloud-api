package jwt

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"example.com/your-api/internal/modules/auth/ports"
)

type Issuer struct {
	issuer string
	aud    string
	kid    string
	ttl    time.Duration
	secret []byte
}

func NewHMACIssuer(issuer, aud, kid, secret string, ttl time.Duration) (*Issuer, error) {
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("jwt secret empty")
	}
	if ttl <= 0 {
		return nil, errors.New("jwt ttl invalid")
	}
	return &Issuer{issuer: issuer, aud: aud, kid: kid, ttl: ttl, secret: []byte(secret)}, nil
}

func (i *Issuer) IssueAccessToken(ctx context.Context, req ports.AccessTokenRequest) (string, time.Time, error) {
	_ = ctx
	if req.AccountID == "" || req.SessionID == "" {
		return "", time.Time{}, errors.New("missing account/session")
	}
	now := time.Now()
	exp := now.Add(i.ttl)

	h := Header{Alg: "HS256", Typ: "JWT", Kid: i.kid}
	p := Payload{
		Iss: i.issuer, Aud: i.aud, Sub: req.AccountID, Sid: req.SessionID, AAL: req.TrustLevel,
		JTI: uuid.NewString(), IAT: now.Unix(), EXP: exp.Unix(),
	}

	hs, err := b64urlEncodeJSON(h)
	if err != nil {
		return "", time.Time{}, err
	}
	ps, err := b64urlEncodeJSON(p)
	if err != nil {
		return "", time.Time{}, err
	}

	input := hs + "." + ps
	sig := signHS256(i.secret, input)
	return fmt.Sprintf("%s.%s", input, sig), exp, nil
}
