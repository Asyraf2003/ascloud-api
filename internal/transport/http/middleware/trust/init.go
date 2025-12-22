package trust

import "github.com/labstack/echo/v4"

func Init(policy string, base int) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if _, ok := Get(c); !ok {
				c.Set(ctxKey, &Context{Policy: policy, Score: base})
			}
			return next(c)
		}
	}
}
