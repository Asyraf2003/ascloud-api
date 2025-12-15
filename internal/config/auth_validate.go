package config

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

func (c AuthConfig) Validate() error {
	if strings.TrimSpace(c.Google.ClientID) == "" {
		return errors.New("missing env: AUTH_GOOGLE_CLIENT_ID")
	}
	if strings.TrimSpace(c.Google.ClientSecret) == "" {
		return errors.New("missing env: AUTH_GOOGLE_CLIENT_SECRET")
	}
	if strings.TrimSpace(c.Google.RedirectURL) == "" {
		return errors.New("missing env: AUTH_GOOGLE_REDIRECT_URL")
	}
	if u, err := url.Parse(c.Google.RedirectURL); err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return fmt.Errorf("invalid AUTH_GOOGLE_REDIRECT_URL: must be http/https")
	}

	if strings.TrimSpace(c.JWT.Secret) == "" {
		return errors.New("missing env: AUTH_JWT_SECRET")
	}
	if len(c.JWT.Secret) < 32 {
		return errors.New("weak AUTH_JWT_SECRET: min length 32")
	}

	if strings.TrimSpace(c.Hash.RefreshPepper) == "" {
		return errors.New("missing env: AUTH_REFRESH_PEPPER")
	}
	if len(c.Hash.RefreshPepper) < 16 {
		return errors.New("weak AUTH_REFRESH_PEPPER: min length 16")
	}

	switch strings.ToLower(strings.TrimSpace(c.Security.CookieSameSite)) {
	case "lax", "strict", "none":
	default:
		return errors.New("invalid COOKIE_SAMESITE: use lax|strict|none")
	}
	return nil
}
