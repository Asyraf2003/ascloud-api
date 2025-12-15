package presenter

import "github.com/labstack/echo/v4"

func AuthSuccess(c echo.Context, status int, payload any) error {
	return c.JSON(status, AuthEnvelope{Auth: payload, Meta: metaFrom(c)})
}
