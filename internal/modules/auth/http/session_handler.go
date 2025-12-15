package http

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"example.com/your-api/internal/config"
	"example.com/your-api/internal/http/presenter"
	"example.com/your-api/internal/modules/auth/domain"
	"example.com/your-api/internal/modules/auth/usecase"
	"example.com/your-api/internal/shared/apperr"
)

type SessionHandler struct {
	flow SessionFlow
	cfg  config.AuthConfig
}

func NewSessionHandler(flow SessionFlow, cfg config.AuthConfig) *SessionHandler {
	return &SessionHandler{flow: flow, cfg: cfg}
}

func (h *SessionHandler) Refresh(c echo.Context) error {
	if c.Request().Method == http.MethodOptions {
		return c.NoContent(http.StatusNoContent)
	}
	if err := requireCSRF(c, h.cfg); err != nil {
		return err
	}

	refresh := readCookie(c, h.cfg.Session.RefreshCookieName)
	out, err := h.flow.Refresh(c.Request().Context(), usecase.RefreshInput{
		RefreshToken: refresh,
		Client:       clientInfoFromEcho(c),
	})
	if err != nil {
		return err
	}

	setNoStore(c)
	if err := setAuthCookies(c, h.cfg, out.RefreshToken, out.RefreshExp, out.CSRFToken); err != nil {
		return apperr.Wrap(err, domain.ErrInternal, http.StatusInternalServerError, "Terjadi kesalahan.")
	}

	return c.JSON(http.StatusOK, presenter.AuthEnvelope{
		Auth: presenter.AuthTokens{
			AccessToken: out.AccessToken, AccessExpiresAt: out.AccessExp.Unix(),
			TrustLevel: out.TrustLevel, StepUpRequired: out.StepUpRequired,
		},
		Meta: &presenter.Meta{},
	})
}

func (h *SessionHandler) Logout(c echo.Context) error {
	if c.Request().Method == http.MethodOptions {
		return c.NoContent(http.StatusNoContent)
	}
	refresh := readCookie(c, h.cfg.Session.RefreshCookieName)
	if refresh != "" {
		if err := requireCSRF(c, h.cfg); err != nil {
			return err
		}
		_ = h.flow.Logout(c.Request().Context(), usecase.LogoutInput{RefreshToken: refresh})
	}
	clearAuthCookies(c, h.cfg)
	setNoStore(c)
	return c.NoContent(http.StatusNoContent)
}
