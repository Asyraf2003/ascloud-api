package middleware

import (
	"crypto/rand"
	"encoding/hex"

	"example.com/your-api/internal/shared/requestid"
	"github.com/labstack/echo/v4"
)

func RequestID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			rid := c.Request().Header.Get(echo.HeaderXRequestID)
			if rid == "" {
				rid = newRequestID()
			}

			// Ensure request/response headers are set.
			c.Request().Header.Set(echo.HeaderXRequestID, rid)
			c.Response().Header().Set(echo.HeaderXRequestID, rid)

			// Also propagate via context for usecases/ports.
			ctx := requestid.With(c.Request().Context(), rid)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}

func newRequestID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
