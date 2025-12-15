package middleware

import "github.com/labstack/echo/v4"

const (
	CtxAccessClaimsKey = "auth.access_claims"
	CtxAccountIDKey    = "auth.account_id"
	CtxSessionIDKey    = "auth.session_id"
	CtxAALKey          = "auth.aal"
)

func AccountID(c echo.Context) (string, bool) {
	v, ok := c.Get(CtxAccountIDKey).(string)
	return v, ok
}

func SessionID(c echo.Context) (string, bool) {
	v, ok := c.Get(CtxSessionIDKey).(string)
	return v, ok
}

func AALLevel(c echo.Context) (string, bool) {
	v, ok := c.Get(CtxAALKey).(string)
	return v, ok
}
