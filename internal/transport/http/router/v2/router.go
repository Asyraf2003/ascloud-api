package v2

import "github.com/labstack/echo/v4"

func Register(e *echo.Echo) {
	_ = e.Group("/v2")
}
