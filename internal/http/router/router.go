package router

import (
	"github.com/labstack/echo/v4"

	auditRouter "example.com/your-api/internal/http/router/audit"
	debugRouter "example.com/your-api/internal/http/router/debug"
	healthRouter "example.com/your-api/internal/http/router/health"
	v1Router "example.com/your-api/internal/http/router/v1"
	v2Router "example.com/your-api/internal/http/router/v2"
	"example.com/your-api/internal/platform/token/jwt"
)

func Register(e *echo.Echo, jwtv *jwt.Verifier) {
	healthRouter.Register(e)
	v1Router.Register(e, jwtv)
	v2Router.Register(e)
	auditRouter.Register(e)
	debugRouter.Register(e)
}
