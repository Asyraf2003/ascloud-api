package router

import (
	"github.com/labstack/echo/v4"

	"example.com/your-api/internal/platform/token/jwt"
	auditRouter "example.com/your-api/internal/transport/http/router/audit"
	debugRouter "example.com/your-api/internal/transport/http/router/debug"
	healthRouter "example.com/your-api/internal/transport/http/router/health"
	v1Router "example.com/your-api/internal/transport/http/router/v1"
	v2Router "example.com/your-api/internal/transport/http/router/v2"
)

func Register(e *echo.Echo, jwtv *jwt.Verifier) {
	healthRouter.Register(e)
	v1Router.Register(e, jwtv)
	v2Router.Register(e)
	auditRouter.Register(e)
	debugRouter.Register(e)
}
