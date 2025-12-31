package auth

import (
	"time"

	"github.com/labstack/echo/v4"

	"example.com/your-api/internal/modules/auth/domain"
	"example.com/your-api/internal/transport/http/middleware/trust"
)

func mwAuthStart() []echo.MiddlewareFunc {
	return []echo.MiddlewareFunc{
		trust.RateLimit("auth_start", 30, time.Minute),
	}
}

func mwAuthCallback() []echo.MiddlewareFunc {
	return []echo.MiddlewareFunc{
		trust.RateLimit("auth_callback", 30, time.Minute),
	}
}

func mwAuthCookieStrict() []echo.MiddlewareFunc {
	pc := getPolicy()

	mws := make([]echo.MiddlewareFunc, 0, 4)

	if pc.requireHTTPS {
		mws = append(mws, trust.RequireHTTPS())
	}

	mws = append(mws,
		trust.RequireOrigin(pc.allowedOrigins),
		trust.RequireCSRFWithCode(pc.csrfCookie, "X-CSRF-Token", domain.ErrCSRFInvalid),
		trust.RateLimit("auth_cookie", 20, time.Minute),
	)

	return mws
}
