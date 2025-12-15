package usecase

import (
	"context"
	"net/http"
	"strings"
	"time"

	"example.com/your-api/internal/modules/auth/domain"
	"example.com/your-api/internal/modules/auth/ports"
	"example.com/your-api/internal/shared/apperr"
)

func (u *GoogleFlow) GoogleStart(ctx context.Context, in GoogleStartInput) (GoogleStartOutput, error) {
	in.Purpose = strings.TrimSpace(in.Purpose)
	if in.Purpose == "" {
		in.Purpose = "login"
	}
	if in.Purpose != "login" && !isStepUpPurpose(in.Purpose) {
		return GoogleStartOutput{}, apperr.New(domain.ErrBadRequest, http.StatusBadRequest, "Permintaan tidak valid.")
	}

	in.RedirectURL = strings.TrimSpace(in.RedirectURL)
	if in.RedirectURL == "" {
		return GoogleStartOutput{}, apperr.New(domain.ErrBadRequest, http.StatusBadRequest, "Permintaan tidak valid.")
	}

	state, err := randB64(32)
	if err != nil {
		return GoogleStartOutput{}, apperr.Wrap(err, domain.ErrInternal, http.StatusInternalServerError, "Terjadi kesalahan.")
	}
	nonce, err := randB64(32)
	if err != nil {
		return GoogleStartOutput{}, apperr.Wrap(err, domain.ErrInternal, http.StatusInternalServerError, "Terjadi kesalahan.")
	}
	verifier, err := randB64(32)
	if err != nil {
		return GoogleStartOutput{}, apperr.Wrap(err, domain.ErrInternal, http.StatusInternalServerError, "Terjadi kesalahan.")
	}

	st := ports.AuthState{Nonce: nonce, CodeVerifier: verifier, Purpose: in.Purpose, CreatedAt: time.Now()}
	if err := u.states.Put(ctx, state, st, u.stateTTL); err != nil {
		return GoogleStartOutput{}, apperr.Wrap(err, domain.ErrInternal, http.StatusInternalServerError, "Terjadi kesalahan.")
	}

	_ = u.audit.Record(ctx, ports.AuditEvent{
		Event: "auth_oidc_start",
		At:    time.Now(),
		Meta:  map[string]any{"provider": "google", "purpose": in.Purpose},
	})

	url := u.oidc.AuthCodeURL(ports.OIDCAuthURLParams{
		State: state, Nonce: nonce, CodeChallenge: pkceChallenge(verifier),
		RedirectURL: in.RedirectURL, Purpose: in.Purpose,
	})
	return GoogleStartOutput{RedirectTo: url, State: state}, nil
}
