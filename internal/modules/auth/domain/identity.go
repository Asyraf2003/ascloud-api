package domain

import "time"

type Identity struct {
	Provider      Provider
	Subject       string
	Email         string
	EmailVerified bool

	// Meta adalah data fleksibel untuk audit/internal.
	// Jangan pernah kirim mentah ke client (policy JSONB).
	Meta map[string]any

	CreatedAt time.Time
	UpdatedAt time.Time
}
