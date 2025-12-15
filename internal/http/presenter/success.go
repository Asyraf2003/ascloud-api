package presenter

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func metaFrom(c echo.Context) *Meta {
	rid := requestID(c)
	if rid == "" {
		return nil
	}
	return &Meta{RequestID: rid}
}

func OK(c echo.Context, data any) error {
	return c.JSON(http.StatusOK, SuccessEnvelope{Data: data, Meta: metaFrom(c)})
}

func Created(c echo.Context, data any) error {
	return c.JSON(http.StatusCreated, SuccessEnvelope{Data: data, Meta: metaFrom(c)})
}

func NoContent(c echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}

func Success(c echo.Context, status int, data any) error {
	return c.JSON(status, SuccessEnvelope{Data: data, Meta: metaFrom(c)})
}
