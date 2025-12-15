package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"example.com/your-api/internal/modules/auth/domain"
	"example.com/your-api/internal/platform/token/jwt"
	"example.com/your-api/internal/shared/apperr"
)

func JWTAuth(v *jwt.Verifier) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// biar preflight CORS gak mati konyol
			if c.Request().Method == http.MethodOptions {
				return next(c)
			}

			h := c.Request().Header.Get("Authorization")
			if h == "" {
				return apperr.New(domain.ErrUnauthorized, http.StatusUnauthorized, "Tidak terautentikasi.")
			}
			if !strings.HasPrefix(h, "Bearer ") {
				return apperr.New(domain.ErrUnauthorized, http.StatusUnauthorized, "Tidak terautentikasi.").
					WithField("reason", "auth_header_not_bearer")
			}

			tok := strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
			cl, err := v.Verify(tok)
			if err != nil {
				return apperr.New(domain.ErrUnauthorized, http.StatusUnauthorized, "Tidak terautentikasi.").
					WithField("reason", "jwt_invalid")
			}

			c.Set(CtxAccessClaimsKey, cl)
			c.Set(CtxAccountIDKey, cl.AccountID)
			c.Set(CtxSessionIDKey, cl.SessionID)
			c.Set(CtxAALKey, cl.TrustLevel)

			return next(c)
		}
	}
}

func AccessClaims(c echo.Context) (jwt.Claims, bool) {
	v, ok := c.Get(CtxAccessClaimsKey).(jwt.Claims)
	return v, ok
}
