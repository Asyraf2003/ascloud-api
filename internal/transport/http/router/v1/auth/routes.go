package auth

import (
	"github.com/labstack/echo/v4"

	authHTTP "example.com/your-api/internal/modules/auth/transport/http"
)

func Register(g *echo.Group) {
	g.GET("/auth/google/start", authHTTP.GoogleStart, mwAuthStart()...)
	g.GET("/auth/google/callback", authHTTP.GoogleCallback, mwAuthCallback()...)

	g.POST("/auth/refresh", authHTTP.Refresh, mwAuthCookieStrict()...)
	g.POST("/auth/logout", authHTTP.Logout, mwAuthCookieStrict()...)
}
