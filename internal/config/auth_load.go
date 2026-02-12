package config

import (
	"strconv"
	"time"
)

func LoadAuth(env string) AuthConfig {
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

	ttlMin := atoi(getenv("AUTH_ACCESS_TTL_MIN", "15"), 15)
	stateMin := atoi(getenv("AUTH_STATE_TTL_MIN", "5"), 5)
	refreshH := atoi(getenv("AUTH_REFRESH_TTL_HOURS", "168"), 168)

	// Dev defaults are allowed. Non-dev must be explicit (fail-fast via Validate()).
	redirectDef := "http://localhost:8080/v1/auth/google/callback"
	originsDef := "http://localhost:8080"
	pepperDef := "dev-pepper-dev-pepper" // >= 16

	if env != "dev" {
		redirectDef = ""
		originsDef = ""
		pepperDef = ""
	}

	return AuthConfig{
		Google: GoogleAuthConfig{
			ClientID:     getenv("AUTH_GOOGLE_CLIENT_ID", ""),
			ClientSecret: getenv("AUTH_GOOGLE_CLIENT_SECRET", ""),
			Issuer:       getenv("AUTH_GOOGLE_ISSUER", "https://accounts.google.com"),
			RedirectURL:  getenv("AUTH_GOOGLE_REDIRECT_URL", redirectDef),
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
			AllowedOrigins: getenvListCSV("AUTH_ALLOWED_ORIGINS", originsDef),
		},
		TTL:  AuthTTLConfig{StateTTL: time.Duration(stateMin) * time.Minute},
		Hash: HashConfig{RefreshPepper: getenv("AUTH_REFRESH_PEPPER", pepperDef)},
	}
}
