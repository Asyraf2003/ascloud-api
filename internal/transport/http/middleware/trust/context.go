package trust

import "github.com/labstack/echo/v4"

const ctxKey = "trust.ctx"

type Context struct {
	Policy  string
	Score   int
	Reasons []string
}

func Get(c echo.Context) (*Context, bool) {
	v, ok := c.Get(ctxKey).(*Context)
	return v, ok
}

func Must(c echo.Context) *Context {
	if v, ok := Get(c); ok {
		return v
	}
	tc := &Context{Policy: "unknown", Score: 0}
	c.Set(ctxKey, tc)
	return tc
}

func Add(c echo.Context, delta int, reason string) {
	tc := Must(c)
	tc.Score += delta
	if reason == "" {
		return
	}
	if len(tc.Reasons) >= 12 {
		return
	}
	tc.Reasons = append(tc.Reasons, reason)
}
