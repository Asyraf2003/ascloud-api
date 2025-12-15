package usecase

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

func randB64(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func pkceChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

func hashRefresh(secret, token string) string {
	m := hmac.New(sha256.New, []byte(secret))
	_, _ = m.Write([]byte(token))
	return base64.RawURLEncoding.EncodeToString(m.Sum(nil))
}

func deref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
