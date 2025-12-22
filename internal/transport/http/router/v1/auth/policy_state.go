package auth

import (
	"sync/atomic"

	"example.com/your-api/internal/config"
)

type policyConfig struct {
	allowedOrigins []string
	csrfCookie     string
}

var policy atomic.Value // *policyConfig

func InitPolicy(cfg config.AuthConfig) {
	pc := &policyConfig{
		allowedOrigins: cfg.Security.AllowedOrigins,
		csrfCookie:     cfg.Session.CSRFCookieName,
	}

	if len(pc.allowedOrigins) == 0 {
		pc.allowedOrigins = []string{"http://localhost:8080"}
	}
	if pc.csrfCookie == "" {
		pc.csrfCookie = "csrf"
	}

	policy.Store(pc)
}

func getPolicy() *policyConfig {
	if v, ok := policy.Load().(*policyConfig); ok && v != nil {
		return v
	}
	return &policyConfig{
		allowedOrigins: []string{"http://localhost:8080"},
		csrfCookie:     "csrf",
	}
}
