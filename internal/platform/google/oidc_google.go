package google

import (
	"context"
	"errors"
	"net/url"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"

	"example.com/your-api/internal/modules/auth/ports"
)

type OIDCConfig struct {
	Issuer       string
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type OIDC struct {
	provider *oidc.Provider
	verifier *oidc.IDTokenVerifier
	oauth    oauth2.Config
}

func NewOIDC(ctx context.Context, cfg OIDCConfig) (*OIDC, error) {
	if cfg.ClientID == "" || cfg.ClientSecret == "" {
		return nil, errors.New("google oidc client id/secret empty")
	}
	p, err := oidc.NewProvider(ctx, cfg.Issuer)
	if err != nil {
		return nil, err
	}
	return &OIDC{
		provider: p,
		verifier: p.Verifier(&oidc.Config{ClientID: cfg.ClientID}),
		oauth: oauth2.Config{
			ClientID: cfg.ClientID, ClientSecret: cfg.ClientSecret,
			Endpoint: p.Endpoint(), RedirectURL: cfg.RedirectURL,
			Scopes: []string{oidc.ScopeOpenID, "email"},
		},
	}, nil
}

func (o *OIDC) AuthCodeURL(p ports.OIDCAuthURLParams) string {
	opts := []oauth2.AuthCodeOption{
		oauth2.SetAuthURLParam("nonce", p.Nonce),
		oauth2.SetAuthURLParam("code_challenge", p.CodeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("prompt", "select_account"),
	}
	if p.RedirectURL != "" {
		u, _ := url.Parse(o.oauth.RedirectURL)
		u.RawQuery = url.Values{"redirect": []string{p.RedirectURL}}.Encode()
	}
	return o.oauth.AuthCodeURL(p.State, opts...)
}

func (o *OIDC) ExchangeAndVerify(ctx context.Context, code, codeVerifier, redirectURL, nonce string) (ports.OIDCClaims, error) {
	tok, err := o.oauth.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", codeVerifier))
	if err != nil {
		return ports.OIDCClaims{}, err
	}
	raw, _ := tok.Extra("id_token").(string)
	if raw == "" {
		return ports.OIDCClaims{}, errors.New("missing id_token")
	}

	idt, err := o.verifier.Verify(ctx, raw)
	if err != nil {
		return ports.OIDCClaims{}, err
	}

	var c struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Nonce         string `json:"nonce"`
		AuthTime      int64  `json:"auth_time"`
	}
	if err := idt.Claims(&c); err != nil {
		return ports.OIDCClaims{}, err
	}
	if nonce != "" && c.Nonce != nonce {
		return ports.OIDCClaims{}, errors.New("nonce mismatch")
	}

	at := time.Unix(c.AuthTime, 0)
	return ports.OIDCClaims{
		Provider: "google", Subject: c.Sub, Email: c.Email,
		EmailVerified: c.EmailVerified, AuthTime: at,
	}, nil
}

var _ ports.OIDCProvider = (*OIDC)(nil)
