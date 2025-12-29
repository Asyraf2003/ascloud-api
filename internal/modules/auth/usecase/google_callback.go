package usecase

import (
	"context"
	"net/http"

	"example.com/your-api/internal/modules/auth/domain"
	"example.com/your-api/internal/shared/apperr"
)

func (u *GoogleFlow) GoogleCallback(ctx context.Context, in GoogleCallbackInput) (GoogleCallbackOutput, error) {
	if in.Code == "" || in.State == "" || in.RedirectURL == "" {
		return GoogleCallbackOutput{}, apperr.New(domain.ErrBadRequest, http.StatusBadRequest, "Permintaan tidak valid.")
	}
	audit0(u, ctx, "", "auth_oidc_callback_attempt", map[string]any{"provider": "google"})

	st, err := u.states.GetDel(ctx, in.State)
	if err != nil {
		audit0(u, ctx, "", "auth_oidc_state_invalid", map[string]any{"provider": "google"})
		return GoogleCallbackOutput{}, apperr.New(domain.ErrOIDCStateInvalid, http.StatusBadRequest, "Permintaan tidak valid.")
	}

	claims, err := u.oidc.ExchangeAndVerify(ctx, in.Code, st.CodeVerifier, in.RedirectURL, st.Nonce)
	if err != nil { /* ... handle err */
	}

	var accID string
	err = u.tx.RunInTx(ctx, func(txCtx context.Context) error {
		id, err := u.resolveAccount(txCtx, claims)
		if err != nil {
			return err
		}
		if err := u.linkIdentity(txCtx, id, claims); err != nil {
			return err
		}
		accID = id
		return nil
	})
	if err != nil {
		return GoogleCallbackOutput{}, err
	}

	// Lanjutkan proses trust & session (di luar transaksi tidak apa-apa karena hanya Read/Token issue)
	trustLevel, stepUp, err := decideTrustAndAAL(u, ctx, accID, st.Purpose, in.Client)
	if err != nil {
		return GoogleCallbackOutput{}, err
	}

	out, err := u.issueSessionAndTokens(ctx, accID, st.Purpose, in.Client, trustLevel)
	if err != nil {
		return GoogleCallbackOutput{}, err
	}
	out.StepUpRequired = stepUp

	audit0(u, ctx, accID, "auth_login_success", map[string]any{
		"provider":        "google",
		"stepup_required": stepUp,
		"trust_level":     trustLevel,
	})
	return out, nil
}
