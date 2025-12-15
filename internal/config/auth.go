package config

import (
	"strconv"
	"time"
)

type AuthConfig struct {
	Google   GoogleAuthConfig
	JWT      JWTConfig
	Session  SessionConfig
	Security CookieSecurityConfig
	TTL      AuthTTLConfig
	Hash     HashConfig
}

type GoogleAuthConfig struct {
	ClientID     string
	ClientSecret string
	Issuer       string
	RedirectURL  string
}

type JWTConfig struct {
	Issuer    string
	Audience  string
	KID       string
	Secret    string
	AccessTTL time.Duration
}

type SessionConfig struct {
	RefreshCookieName string
	CSRFCookieName    string
	RefreshTTL        time.Duration
}

type CookieSecurityConfig struct {
	CookieDomain   string
	CookieSecure   bool
	CookieSameSite string // strict|lax|none
}

type AuthTTLConfig struct {
	StateTTL time.Duration
}

type HashConfig struct {
	RefreshPepper string
}

func LoadAuth() AuthConfig {
	atoi := func(s string, def int) int {
		n, err := strconv.Atoi(s)
		if err != nil {
			return def
		}
		return n
	}
	atob := func(s string, def bool) bool {
		b, err := strconv.ParseBool(s)
		if err != nil {
			return def
		}
		return b
	}

	ttlMin := atoi(getenv("AUTH_ACCESS_TTL_MIN", "30"), 30)
	stateMin := atoi(getenv("AUTH_STATE_TTL_MIN", "5"), 5)
	refreshH := atoi(getenv("AUTH_REFRESH_TTL_HOURS", "168"), 168)

	return AuthConfig{
		Google: GoogleAuthConfig{
			ClientID:     getenv("AUTH_GOOGLE_CLIENT_ID", ""),
			ClientSecret: getenv("AUTH_GOOGLE_CLIENT_SECRET", ""),
			Issuer:       getenv("AUTH_GOOGLE_ISSUER", "https://accounts.google.com"),
			RedirectURL:  getenv("AUTH_GOOGLE_REDIRECT_URL", "http://localhost:8080/v1/auth/google/callback"),
		},
		JWT: JWTConfig{
			Issuer:    getenv("AUTH_JWT_ISSUER", "example.com/your-api"),
			Audience:  getenv("AUTH_JWT_AUDIENCE", "your-api"),
			KID:       getenv("AUTH_JWT_KID", "dev"),
			Secret:    getenv("AUTH_JWT_SECRET", ""),
			AccessTTL: time.Duration(ttlMin) * time.Minute,
		},
		Session: SessionConfig{
			RefreshCookieName: getenv("AUTH_REFRESH_COOKIE", "refresh"),
			CSRFCookieName:    getenv("AUTH_CSRF_COOKIE", "csrf"),
			RefreshTTL:        time.Duration(refreshH) * time.Hour,
		},
		Security: CookieSecurityConfig{
			CookieDomain:   getenv("COOKIE_DOMAIN", ""),
			CookieSecure:   atob(getenv("COOKIE_SECURE", "false"), false),
			CookieSameSite: getenv("COOKIE_SAMESITE", "lax"),
		},
		TTL:  AuthTTLConfig{StateTTL: time.Duration(stateMin) * time.Minute},
		Hash: HashConfig{RefreshPepper: getenv("AUTH_REFRESH_PEPPER", "dev-pepper")},
	}
}
