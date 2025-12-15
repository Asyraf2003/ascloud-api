package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
)

var errInvalidToken = errors.New("invalid token")

func b64urlEncodeJSON(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func b64urlDecodeJSON(seg string, dst any) error {
	b, err := base64.RawURLEncoding.DecodeString(seg)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dst)
}

func signHS256(secret []byte, input string) string {
	m := hmac.New(sha256.New, secret)
	_, _ = m.Write([]byte(input))
	return base64.RawURLEncoding.EncodeToString(m.Sum(nil))
}

func verifyHS256(secret []byte, input, sig string) bool {
	want := signHS256(secret, input)
	return hmac.Equal([]byte(want), []byte(sig))
}
