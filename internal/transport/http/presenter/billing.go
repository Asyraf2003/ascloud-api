package presenter

import "github.com/labstack/echo/v4"

func BillingSuccess(c echo.Context, status int, payload any) error {
	return c.JSON(status, BillingEnvelope{Billing: payload, Meta: metaFrom(c)})
}
