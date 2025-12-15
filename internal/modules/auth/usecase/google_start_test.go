package usecase

import (
	"context"
	"testing"
	"time"

	"example.com/your-api/internal/modules/auth/ports"
)

type fakeOIDC struct{}

func (f fakeOIDC) AuthCodeURL(p ports.OIDCAuthURLParams) string {
	return "https://google/auth?state=" + p.State
}
func (f fakeOIDC) ExchangeAndVerify(ctx context.Context, code, v, r, n string) (ports.OIDCClaims, error) {
	return ports.OIDCClaims{}, nil
}

func TestGoogleStart_OK(t *testing.T) {
	u, err := NewGoogleFlow(fakeOIDC{}, memState(), memSess(), memIDs(), memAcc(), memTok(), fakeTrust{}, fakeAudit{},
		5*time.Minute, 7*24*time.Hour, "pepper")
	if err != nil {
		t.Fatal(err)
	}
	out, err := u.GoogleStart(context.Background(), GoogleStartInput{Purpose: "login", RedirectURL: "http://cb"})
	if err != nil || out.State == "" || out.RedirectTo == "" {
		t.Fatalf("bad out err=%v out=%+v", err, out)
	}
}
