package domain

import "errors"

// Error domain auth (dipakai usecase, bukan vendor error).
var (
	ErrSessionNotFound    = errors.New("auth session not found")
	ErrRefreshTokenReused = errors.New("refresh token reused or invalid (possible theft)")
	ErrSessionExpired     = errors.New("auth session expired")
	ErrSessionRevoked     = errors.New("auth session revoked")
)
