package http

import (
	"context"

	"example.com/your-api/internal/modules/auth/usecase"
)

type SessionFlow interface {
	Refresh(ctx context.Context, in usecase.RefreshInput) (usecase.RefreshOutput, error)
	Logout(ctx context.Context, in usecase.LogoutInput) error
}
