package v1

import (
	"github.com/labstack/echo/v4"

	"example.com/your-api/internal/platform/token/jwt"
	"example.com/your-api/internal/transport/http/middleware"

	accountPkg "example.com/your-api/internal/transport/http/router/v1/account"
	authPkg "example.com/your-api/internal/transport/http/router/v1/auth"
	billingPkg "example.com/your-api/internal/transport/http/router/v1/billing"
	domainMgmtPkg "example.com/your-api/internal/transport/http/router/v1/domainmgmt"
	healthPkg "example.com/your-api/internal/transport/http/router/v1/health"
	hostingPkg "example.com/your-api/internal/transport/http/router/v1/hosting"
	mePkg "example.com/your-api/internal/transport/http/router/v1/me"
	trustPkg "example.com/your-api/internal/transport/http/router/v1/trust"
)

func Register(e *echo.Echo, jwtv *jwt.Verifier) {
	base := e.Group("/v1")

	public := base.Group("")
	healthPkg.Register(public)
	authPkg.Register(public)

	protected := base.Group("")
	protected.Use(middleware.JWTAuth(jwtv))

	mePkg.Register(protected) // sanity endpoint
	accountPkg.Register(protected)
	hostingPkg.Register(protected)
	domainMgmtPkg.Register(protected)
	trustPkg.Register(protected)

	billingPkg.Register(protected) // kalau masih cangkang, keep dulu
}
