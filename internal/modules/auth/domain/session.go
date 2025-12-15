package domain

import "time"

// Session = sesi login server-side.
// Refresh token tidak pernah disimpan plain, hanya hash-nya.
type Session struct {
	ID        string
	UserID    string
	ProjectID *string

	RefreshTokenHash string

	DeviceID      string
	UserAgentHash string
	IPPrefix      *string

	CreatedAt time.Time
	ExpiresAt time.Time
	RevokedAt *time.Time

	Meta map[string]any
}
