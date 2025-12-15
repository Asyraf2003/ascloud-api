package ports

import (
	"context"

	"github.com/google/uuid"
)

type AccountInput struct {
	Email string
	Meta  map[string]any
}

type AccountService interface {
	Create(ctx context.Context, in AccountInput) (uuid.UUID, error)
}
