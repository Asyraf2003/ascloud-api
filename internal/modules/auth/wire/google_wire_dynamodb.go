package wire

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"example.com/your-api/internal/config"
	"example.com/your-api/internal/modules/auth/ports"
	authddb "example.com/your-api/internal/modules/auth/store/dynamodb"
	authHTTP "example.com/your-api/internal/modules/auth/transport/http"
	"example.com/your-api/internal/modules/auth/usecase"
	google "example.com/your-api/internal/platform/google"
	memstate "example.com/your-api/internal/platform/state/memory"
	jwt "example.com/your-api/internal/platform/token/jwt"
)

type allowAllTrust struct{}

func (a allowAllTrust) Evaluate(ctx context.Context, s ports.TrustSignals) (ports.TrustDecision, error) {
	return ports.TrustDecision{Allow: true}, nil
}

func WireAuthGoogle(ddb *dynamodb.Client, cfg config.AuthConfig) error {
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

	tables := authddb.TableNamesFromEnv("")
	flow, err := usecase.NewGoogleFlow(
		oidc,
		memstate.NewAuthStateStore(),
		authddb.NewSessionStore(ddb, tables.Sessions),
		authddb.NewIdentityRepo(ddb, tables.Identities),
		authddb.NewAccountService(ddb, tables.Accounts),
		issuer,
		allowAllTrust{},
		authddb.NewAuditSink(ddb, tables.AuditEvents),
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
