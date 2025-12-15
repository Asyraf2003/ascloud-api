package health

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func Register(g *echo.Group) {
	g.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]any{"status": "ok", "v": "v1"})
	})
}
