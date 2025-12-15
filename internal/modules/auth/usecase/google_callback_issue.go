package usecase

import (
	"context"
	"net/http"
	"time"

	"example.com/your-api/internal/modules/auth/domain"
	"example.com/your-api/internal/modules/auth/ports"
	"example.com/your-api/internal/shared/apperr"
)

func (u *GoogleFlow) issueSessionAndTokens(
	ctx context.Context,
	accountID, purpose string,
	client ClientInfo,
	trustLevel string,
) (GoogleCallbackOutput, error) {
	refresh, err := randB64(32)
	if err != nil {
		return GoogleCallbackOutput{}, apperr.Wrap(err, domain.ErrInternal, http.StatusInternalServerError, "Terjadi kesalahan.")
	}
	csrf, err := randB64(24)
	if err != nil {
		return GoogleCallbackOutput{}, apperr.Wrap(err, domain.ErrInternal, http.StatusInternalServerError, "Terjadi kesalahan.")
	}

	now := time.Now()
	refreshExp := now.Add(u.refreshTTL)
	sess, err := u.sessions.Create(ctx, domain.Session{
		UserID:           accountID,
		RefreshTokenHash: hashRefresh(u.hashSecret, refresh),
		DeviceID:         client.DeviceID,
		UserAgentHash:    client.UserAgentHash,
		IPPrefix:         client.IPPrefix,
		ExpiresAt:        refreshExp,
		Meta:             map[string]any{"purpose": purpose, "aal": trustLevel},
	})
	if err != nil {
		return GoogleCallbackOutput{}, apperr.Wrap(err, domain.ErrInternal, http.StatusInternalServerError, "Terjadi kesalahan.")
	}

	at, exp, err := u.tokens.IssueAccessToken(ctx, ports.AccessTokenRequest{
		AccountID:  accountID,
		SessionID:  sess.ID,
		TrustLevel: trustLevel,
	})
	if err != nil {
		return GoogleCallbackOutput{}, apperr.Wrap(err, domain.ErrInternal, http.StatusInternalServerError, "Terjadi kesalahan.")
	}

	return GoogleCallbackOutput{
		AccountID:      accountID,
		SessionID:      sess.ID,
		AccessToken:    at,
		AccessExp:      exp,
		RefreshToken:   refresh,
		RefreshExp:     refreshExp,
		CSRFToken:      csrf,
		TrustLevel:     trustLevel,
		StepUpRequired: false, // diset di caller
	}, nil
}
