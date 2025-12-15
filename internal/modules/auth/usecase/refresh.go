package usecase

import (
	"context"
	"net/http"
	"time"

	"example.com/your-api/internal/modules/auth/domain"
	"example.com/your-api/internal/modules/auth/ports"
	"example.com/your-api/internal/shared/apperr"
)

func (u *GoogleFlow) Refresh(ctx context.Context, in RefreshInput) (RefreshOutput, error) {
	if in.RefreshToken == "" {
		return RefreshOutput{}, apperr.New(domain.ErrRefreshMissing, http.StatusUnauthorized, "Tidak terautentikasi.")
	}

	oldHash := hashRefresh(u.hashSecret, in.RefreshToken)
	sess, err := u.sessions.GetByRefreshTokenHash(ctx, oldHash)
	if err != nil || sess.RevokedAt != nil || time.Now().After(sess.ExpiresAt) {
		return RefreshOutput{}, apperr.New(domain.ErrUnauthorized, http.StatusUnauthorized, "Tidak terautentikasi.")
	}

	dec, derr := trustDecision(u, ctx, sess.UserID, "refresh", in.Client)
	if derr == nil && !dec.Allow {
		return RefreshOutput{}, apperr.New(domain.ErrForbidden, http.StatusForbidden, "Tidak punya akses.").
			WithField("reason", dec.Reason)
	}
	stepUp := derr != nil || dec.RequireStepUp

	newRefresh, err := randB64(32)
	if err != nil {
		return RefreshOutput{}, apperr.Wrap(err, domain.ErrInternal, http.StatusInternalServerError, "Terjadi kesalahan.")
	}
	newCSRF, err := randB64(24)
	if err != nil {
		return RefreshOutput{}, apperr.Wrap(err, domain.ErrInternal, http.StatusInternalServerError, "Terjadi kesalahan.")
	}

	refreshExp := time.Now().Add(u.refreshTTL)
	newHash := hashRefresh(u.hashSecret, newRefresh)

	if err := u.sessions.RotateRefreshTokenHash(ctx, sess.ID, oldHash, newHash, refreshExp); err != nil {
		_ = u.sessions.Revoke(ctx, sess.ID, time.Now())
		return RefreshOutput{}, apperr.New(domain.ErrRefreshReused, http.StatusUnauthorized, "Tidak terautentikasi.")
	}

	trust := aalFromSession(sess)
	if stepUp {
		trust = "aal1"
	}

	at, atExp, err := u.tokens.IssueAccessToken(ctx, ports.AccessTokenRequest{
		AccountID: sess.UserID, SessionID: sess.ID, TrustLevel: trust,
	})
	if err != nil {
		return RefreshOutput{}, apperr.Wrap(err, domain.ErrInternal, http.StatusInternalServerError, "Terjadi kesalahan.")
	}

	audit0(u, ctx, sess.UserID, "auth_refresh_success", map[string]any{"stepup_required": stepUp})
	return RefreshOutput{
		AccountID: sess.UserID, SessionID: sess.ID,
		AccessToken: at, AccessExp: atExp,
		RefreshToken: newRefresh, RefreshExp: refreshExp,
		CSRFToken:  newCSRF,
		TrustLevel: trust, StepUpRequired: stepUp,
	}, nil
}
