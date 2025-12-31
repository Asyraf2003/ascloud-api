package middleware

import (
	"sync/atomic"

	"example.com/your-api/internal/shared/authn"
)

var accessTokenVerifier atomic.Value // authn.AccessTokenVerifier

func SetAccessTokenVerifier(v authn.AccessTokenVerifier) {
	accessTokenVerifier.Store(v)
}

func AccessTokenVerifier() (authn.AccessTokenVerifier, bool) {
	v, ok := accessTokenVerifier.Load().(authn.AccessTokenVerifier)
	return v, ok && v != nil
}

func MustAccessTokenVerifier() authn.AccessTokenVerifier {
	v, ok := AccessTokenVerifier()
	if !ok {
		panic("access token verifier not set")
	}
	return v
}
