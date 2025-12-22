package v1

import (
	"time"

	"github.com/labstack/echo/v4"

	"example.com/your-api/internal/platform/token/jwt"
	"example.com/your-api/internal/transport/http/middleware"
	"example.com/your-api/internal/transport/http/middleware/trust"

	accountPkg "example.com/your-api/internal/transport/http/router/v1/account"
	authPkg "example.com/your-api/internal/transport/http/router/v1/auth"
	billingPkg "example.com/your-api/internal/transport/http/router/v1/billing"
	domainMgmtPkg "example.com/your-api/internal/transport/http/router/v1/domains"
	healthPkg "example.com/your-api/internal/transport/http/router/v1/health"
	hostingPkg "example.com/your-api/internal/transport/http/router/v1/hosting"
	mePkg "example.com/your-api/internal/transport/http/router/v1/me"
	trustPkg "example.com/your-api/internal/transport/http/router/v1/trust"
)

func Register(e *echo.Echo, jwtv *jwt.Verifier) {
	base := e.Group("/v1")

	pub := base.Group("")
	healthPkg.Register(pub)

	authG := base.Group("")
	authG.Use(trust.Init("auth", 50))
	authG.Use(trust.RequireHTTPS())
	authG.Use(trust.UserAgentScore())
	authG.Use(trust.RateLimit("auth_group", 120, time.Minute))
	authPkg.Register(authG)

	protected := base.Group("")
	protected.Use(trust.Init("api", 50))
	protected.Use(trust.RequireHTTPS())
	protected.Use(trust.UserAgentScore())
	protected.Use(middleware.JWTAuth(jwtv))
	protected.Use(trust.ScoreFromAAL())
	protected.Use(trust.Enforce(trust.Thresholds{Allow: 75, StepUp: 50}))

	mePkg.Register(protected)
	accountPkg.Register(protected)
	hostingPkg.Register(protected)
	domainMgmtPkg.Register(protected)
	trustPkg.Register(protected)
	billingPkg.Register(protected)
}
