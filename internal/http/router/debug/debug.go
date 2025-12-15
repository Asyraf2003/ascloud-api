package debug

import (
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
)

func Register(e *echo.Echo) {
	if os.Getenv("DEBUG_ROUTES") != "1" {
		return
	}

	g := e.Group("/__debug")
	g.GET("/ping", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]any{"ok": true})
	})
}
