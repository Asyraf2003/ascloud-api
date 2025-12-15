package http

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"example.com/your-api/internal/config"
	"example.com/your-api/internal/modules/auth/domain"
	"example.com/your-api/internal/modules/auth/usecase"
	"example.com/your-api/internal/shared/apperr"
	"example.com/your-api/internal/transport/http/presenter"
)

type GoogleHandler struct {
	flow GoogleFlow
	cfg  config.AuthConfig
}

func NewGoogleHandler(flow GoogleFlow, cfg config.AuthConfig) *GoogleHandler {
	return &GoogleHandler{flow: flow, cfg: cfg}
}

func (h *GoogleHandler) Start(c echo.Context) error {
	purpose := c.QueryParam("purpose")
	out, err := h.flow.GoogleStart(c.Request().Context(), usecase.GoogleStartInput{
		Purpose: purpose, RedirectURL: h.cfg.Google.RedirectURL,
	})
	if err != nil {
		return err
	}
	return c.Redirect(http.StatusFound, out.RedirectTo)
}

func (h *GoogleHandler) Callback(c echo.Context) error {
	if c.QueryParam("error") != "" {
		return apperr.New(domain.ErrUnauthorized, http.StatusUnauthorized, "Tidak terautentikasi.")
	}

	out, err := h.flow.GoogleCallback(c.Request().Context(), usecase.GoogleCallbackInput{
		Code: c.QueryParam("code"), State: c.QueryParam("state"),
		RedirectURL: h.cfg.Google.RedirectURL,
		Client:      clientInfoFromEcho(c),
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

func setNoStore(c echo.Context) {
	c.Response().Header().Set("Cache-Control", "no-store")
	c.Response().Header().Set("Pragma", "no-cache")
	c.Response().Header().Set("Expires", time.Unix(0, 0).UTC().Format(time.RFC1123))
}
