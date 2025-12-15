package usecase

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"

	"example.com/your-api/internal/modules/auth/domain"
	"example.com/your-api/internal/modules/auth/ports"
	"example.com/your-api/internal/shared/apperr"
)

func (u *GoogleFlow) resolveAccount(ctx context.Context, c ports.OIDCClaims) (string, error) {
	accID, ok, err := u.ids.FindAccountIDByIdentity(ctx, c.Provider, c.Subject)
	if err != nil {
		return "", apperr.Wrap(err, domain.ErrInternal, http.StatusInternalServerError, "Terjadi kesalahan.")
	}
	if ok {
		return accID.String(), nil
	}
	id, err := u.accounts.Create(ctx, ports.AccountInput{Email: c.Email, Meta: map[string]any{"source": "google"}})
	if err != nil {
		return "", apperr.Wrap(err, domain.ErrInternal, http.StatusInternalServerError, "Terjadi kesalahan.")
	}
	return id.String(), nil
}

func (u *GoogleFlow) linkIdentity(ctx context.Context, accountID string, c ports.OIDCClaims) error {
	acc, err := uuid.Parse(accountID)
	if err != nil {
		return apperr.New(domain.ErrInternal, http.StatusInternalServerError, "Terjadi kesalahan.")
	}
	if err := u.ids.UpsertIdentity(ctx, acc, c.Provider, c.Subject, c.Email, c.EmailVerified, map[string]any{"auth_time": time.Now().Unix()}); err != nil {
		return apperr.New(domain.ErrForbidden, http.StatusForbidden, "Tidak punya akses.")
	}
	return nil
}
