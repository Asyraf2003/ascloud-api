package http

import (
	"context"

	"example.com/your-api/internal/modules/auth/usecase"
)

type GoogleFlow interface {
	GoogleStart(ctx context.Context, in usecase.GoogleStartInput) (usecase.GoogleStartOutput, error)
	GoogleCallback(ctx context.Context, in usecase.GoogleCallbackInput) (usecase.GoogleCallbackOutput, error)
}
