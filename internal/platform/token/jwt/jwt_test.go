package jwt

import (
	"context"
	"testing"
	"time"

	"example.com/your-api/internal/modules/auth/ports"
)

func TestIssueAndVerify(t *testing.T) {
	iss, err := NewHMACIssuer("iss", "aud", "kid", "secretsecret", 30*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	vf, err := NewHMACVerifier("iss", "aud", "secretsecret")
	if err != nil {
		t.Fatal(err)
	}

	tok, _, err := iss.IssueAccessToken(context.Background(), ports.AccessTokenRequest{
		AccountID: "acc", SessionID: "sid", TrustLevel: "aal1",
	})
	if err != nil {
		t.Fatal(err)
	}
	c, err := vf.Verify(tok)
	if err != nil {
		t.Fatal(err)
	}
	if c.AccountID != "acc" || c.SessionID != "sid" || c.TrustLevel != "aal1" {
		t.Fatal("claims mismatch")
	}
}

func TestExpired(t *testing.T) {
	iss, err := NewHMACIssuer("iss", "aud", "kid", "secretsecret", 1*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	vf, err := NewHMACVerifier("iss", "aud", "secretsecret")
	if err != nil {
		t.Fatal(err)
	}

	tok, _, err := iss.IssueAccessToken(context.Background(), ports.AccessTokenRequest{
		AccountID: "acc", SessionID: "sid", TrustLevel: "aal1",
	})
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(3 * time.Millisecond)

	if _, err := vf.Verify(tok); err == nil {
		t.Fatal("expected expired error")
	}
}
