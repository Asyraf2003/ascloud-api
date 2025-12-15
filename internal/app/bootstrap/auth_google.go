package bootstrap

import (
	"context"
	"database/sql"
	"time"

	authPkg "example.com/your-api/internal/platform/datastore/postgres/auth"

	"example.com/your-api/internal/config"
	"example.com/your-api/internal/modules/auth/ports"
	authHTTP "example.com/your-api/internal/modules/auth/transport/http"
	"example.com/your-api/internal/modules/auth/usecase"
	"example.com/your-api/internal/platform/google"
	memstate "example.com/your-api/internal/platform/state/memory"
	jwt "example.com/your-api/internal/platform/token/jwt"
)

type allowAllTrust struct{}

func (a allowAllTrust) Evaluate(ctx context.Context, s ports.TrustSignals) (ports.TrustDecision, error) {
	return ports.TrustDecision{Allow: true}, nil
}

func WireAuthGoogle(db *sql.DB, cfg config.AuthConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	oidc, err := google.NewOIDC(ctx, google.OIDCConfig{
		Issuer: cfg.Google.Issuer, ClientID: cfg.Google.ClientID, ClientSecret: cfg.Google.ClientSecret, RedirectURL: cfg.Google.RedirectURL,
	})
	if err != nil {
		return err
	}

	issuer, err := jwt.NewHMACIssuer(cfg.JWT.Issuer, cfg.JWT.Audience, cfg.JWT.KID, cfg.JWT.Secret, cfg.JWT.AccessTTL)
	if err != nil {
		return err
	}

	flow, err := usecase.NewGoogleFlow(
		oidc,
		memstate.NewAuthStateStore(),
		authPkg.NewSessionStorePostgres(db),
		authPkg.NewAuthIdentityRepo(db),
		authPkg.NewAuthAccountService(db),
		issuer,
		allowAllTrust{},
		authPkg.NewAuthAuditSink(db),
		cfg.TTL.StateTTL,
		cfg.Session.RefreshTTL,
		cfg.Hash.RefreshPepper,
	)
	if err != nil {
		return err
	}

	authHTTP.SetGoogleHandler(authHTTP.NewGoogleHandler(flow, cfg))
	authHTTP.SetSessionHandler(authHTTP.NewSessionHandler(flow, cfg))
	return nil
}
