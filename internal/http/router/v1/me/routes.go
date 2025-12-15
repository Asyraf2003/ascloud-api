package me

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"example.com/your-api/internal/http/middleware"
)

func Register(g *echo.Group) {
	g.GET("/me", func(c echo.Context) error {
		aid, _ := middleware.AccountID(c)
		sid, _ := middleware.SessionID(c)
		aal, _ := middleware.AALLevel(c)

		return c.JSON(http.StatusOK, map[string]any{
			"account_id": aid,
			"session_id": sid,
			"aal":        aal,
		})
	})
}
