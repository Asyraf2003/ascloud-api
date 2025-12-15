package ports

import (
	"context"

	"github.com/google/uuid"
)

type IdentityRepository interface {
	FindAccountIDByIdentity(ctx context.Context, provider, subject string) (uuid.UUID, bool, error)

	UpsertIdentity(ctx context.Context, accountID uuid.UUID, provider, subject, email string, emailVerified bool, meta map[string]any) error
}
