package usecase

import (
	"context"
	"net/http"

	"example.com/your-api/internal/modules/auth/domain"
	"example.com/your-api/internal/shared/apperr"
)

func isStepUpPurpose(p string) bool {
	return p == string(domain.PurposeStepUp) || p == "stepup"
}

func decideTrustAndAAL(u *GoogleFlow, ctx context.Context, accountID, purpose string, c ClientInfo) (string, bool, error) {
	dec, derr := trustDecision(u, ctx, accountID, purpose, c)
	if derr == nil && !dec.Allow {
		audit0(u, ctx, accountID, "auth_trust_denied", map[string]any{
			"reason":  dec.Reason,
			"purpose": purpose,
		})
		return "", false, apperr.New(domain.ErrForbidden, http.StatusForbidden, "Tidak punya akses.").
			WithField("reason", dec.Reason)
	}

	stepUp := derr != nil || dec.RequireStepUp
	trustLevel := "aal1"

	if isStepUpPurpose(purpose) {
		if derr == nil && dec.Allow && !dec.RequireStepUp {
			trustLevel = "aal2"
			stepUp = false
		} else {
			trustLevel = "aal1"
			stepUp = true
		}
	}

	return trustLevel, stepUp, nil
}
