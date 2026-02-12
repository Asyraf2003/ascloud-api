package auth

import (
	"sync/atomic"

	"example.com/your-api/internal/config"
)

type policyConfig struct {
	allowedOrigins []string
	csrfCookie     string
	requireHTTPS   bool
}

var policy atomic.Value // *policyConfig

func InitPolicy(cfg config.AuthConfig) {
	pc := &policyConfig{
		allowedOrigins: cfg.Security.AllowedOrigins,
		csrfCookie:     cfg.Session.CSRFCookieName,
		requireHTTPS:   cfg.Security.CookieSecure,
	}

	if pc.csrfCookie == "" {
		pc.csrfCookie = "csrf"
	}

	policy.Store(pc)
}

func RequireHTTPS() bool {
	return getPolicy().requireHTTPS
}

func getPolicy() *policyConfig {
	if v, ok := policy.Load().(*policyConfig); ok && v != nil {
		return v
	}
	// Startup should call InitPolicy(). If it didn't, we prefer failing closed.
	return &policyConfig{
		allowedOrigins: nil,
		csrfCookie:     "csrf",
		requireHTTPS:   false,
	}
}
