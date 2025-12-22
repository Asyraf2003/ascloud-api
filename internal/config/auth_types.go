package config

import "time"

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
	AllowedOrigins []string
}

type AuthTTLConfig struct {
	StateTTL time.Duration
}

type HashConfig struct {
	RefreshPepper string
}
