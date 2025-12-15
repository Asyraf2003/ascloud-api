package auth

import (
	"github.com/labstack/echo/v4"

	authHTTP "example.com/your-api/internal/modules/auth/http"
)

func Register(g *echo.Group) {
	g.GET("/auth/google/start", authHTTP.GoogleStart)
	g.GET("/auth/google/callback", authHTTP.GoogleCallback)
	g.POST("/auth/refresh", authHTTP.Refresh)
	g.POST("/auth/logout", authHTTP.Logout)
}
