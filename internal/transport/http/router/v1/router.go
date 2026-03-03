package v1

import (
	"time"

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

	"github.com/labstack/echo/v4"
)

func Register(e *echo.Echo) {
	base := e.Group("/v1")

	// --- Public Routes ---
	pub := base.Group("")
	healthPkg.Register(pub)

	// --- Auth Routes ---
	authG := base.Group("")
	authG.Use(trust.Init("auth", 50))
	if authPkg.RequireHTTPS() {
		authG.Use(trust.RequireHTTPS())
	}
	authG.Use(trust.UserAgentScore())
	authG.Use(trust.RateLimit("auth_group", 120, time.Minute))
	authPkg.Register(authG)

	// --- Protected Base (Common Middlewares) ---
	protectedBase := base.Group("")
	protectedBase.Use(trust.Init("api", 50))
	if authPkg.RequireHTTPS() {
		protectedBase.Use(trust.RequireHTTPS())
	}
	protectedBase.Use(trust.UserAgentScore())
	protectedBase.Use(middleware.JWTAuth())
	protectedBase.Use(trust.ScoreFromAAL())

	// --- Low-Risk Routes (AAL1 + MVP flows) ---
	low := protectedBase.Group("")
	mePkg.Register(low)
	hostingPkg.Register(low)

	// --- Default Protected Routes (Stricter) ---
	def := protectedBase.Group("")
	def.Use(trust.Enforce(trust.Thresholds{
		Allow:  75,
		StepUp: 50,
	}))
	accountPkg.Register(def)
	trustPkg.Register(def)

	// --- High-Risk Routes (Billing/Domains) ---
	high := protectedBase.Group("")
	high.Use(trust.Enforce(trust.Thresholds{
		Allow:  75,
		StepUp: 50,
	}))
	domainMgmtPkg.Register(high)
	billingPkg.Register(high)
}
