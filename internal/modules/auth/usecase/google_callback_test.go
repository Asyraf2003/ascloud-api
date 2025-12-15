package usecase

import (
	"context"
	"testing"
	"time"

	"example.com/your-api/internal/modules/auth/ports"
)

type oidcOK struct{}

func (o oidcOK) AuthCodeURL(p ports.OIDCAuthURLParams) string { return "u" }
func (o oidcOK) ExchangeAndVerify(ctx context.Context, code, v, r, n string) (ports.OIDCClaims, error) {
	return ports.OIDCClaims{Provider: "google", Subject: "sub", Email: "a@b.com", EmailVerified: true, AuthTime: time.Now()}, nil
}

func TestGoogleCallback_OK(t *testing.T) {
	st := memState()
	_ = st.Put(context.Background(), "state", ports.AuthState{Nonce: "n", CodeVerifier: "v", Purpose: "login", CreatedAt: time.Now()}, time.Minute)

	u, err := NewGoogleFlow(oidcOK{}, st, memSess(), memIDs(), memAcc(), memTok(), fakeTrust{}, fakeAudit{}, 5*time.Minute, 7*24*time.Hour, "pepper")
	if err != nil {
		t.Fatal(err)
	}

	out, err := u.GoogleCallback(context.Background(), GoogleCallbackInput{
		Code: "c", State: "state", RedirectURL: "http://cb",
		Client: ClientInfo{DeviceID: "d", UserAgentHash: "ua"},
	})
	if err != nil || out.AccessToken == "" || out.RefreshToken == "" || out.CSRFToken == "" {
		t.Fatalf("bad out err=%v out=%+v", err, out)
	}
}
