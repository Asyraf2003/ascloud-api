package trust

import (
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"

	"example.com/your-api/internal/shared/apperr"
	"example.com/your-api/internal/transport/http/middleware"
)

type entry struct {
	n     int
	reset time.Time
}

func RateLimit(name string, limit int, window time.Duration) echo.MiddlewareFunc {
	var mu sync.Mutex
	m := map[string]*entry{}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().Method == http.MethodOptions {
				return next(c)
			}

			key := name + ":" + clientKey(c)
			now := time.Now()

			mu.Lock()
			e := m[key]
			if e == nil || now.After(e.reset) {
				e = &entry{n: 0, reset: now.Add(window)}
				m[key] = e
			}
			e.n++
			n := e.n
			mu.Unlock()

			if n > limit {
				return apperr.New("RATE_LIMITED", http.StatusTooManyRequests, "Terlalu banyak permintaan.").
					WithField("reason", "rate_limited").
					WithField("bucket", name)
			}

			return next(c)
		}
	}
}

func clientKey(c echo.Context) string {
	if acc, ok := middleware.AccountID(c); ok && acc != "" {
		return "acc:" + acc
	}
	return "ip:" + c.RealIP()
}
