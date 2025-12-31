package router

import (
	"github.com/labstack/echo/v4"

	"example.com/your-api/internal/shared/authn"
	"example.com/your-api/internal/transport/http/middleware"
	auditRouter "example.com/your-api/internal/transport/http/router/audit"
	debugRouter "example.com/your-api/internal/transport/http/router/debug"
	healthRouter "example.com/your-api/internal/transport/http/router/health"
	v1Router "example.com/your-api/internal/transport/http/router/v1"
	v2Router "example.com/your-api/internal/transport/http/router/v2"
)

func SetAccessTokenVerifier(v authn.AccessTokenVerifier) {
	middleware.SetAccessTokenVerifier(v)
}

func Register(e *echo.Echo) {
	healthRouter.Register(e)
	v1Router.Register(e)
	v2Router.Register(e)
	auditRouter.Register(e)
	debugRouter.Register(e)
}
