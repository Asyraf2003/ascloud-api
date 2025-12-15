package usecase

import (
	"context"
	"testing"
	"time"

	"example.com/your-api/internal/modules/auth/ports"
)

type trustDec struct {
	dec ports.TrustDecision
	err error
}

func (t trustDec) Evaluate(ctx context.Context, s ports.TrustSignals) (ports.TrustDecision, error) {
	return t.dec, t.err
}

func TestGoogleCallback_StepUp_ElevatesToAAL2_WhenTrustOK(t *testing.T) {
	st := memState()
	_ = st.Put(context.Background(), "state",
		ports.AuthState{Nonce: "n", CodeVerifier: "v", Purpose: "stepup", CreatedAt: time.Now()},
		time.Minute,
	)

	u, _ := NewGoogleFlow(oidcOK{}, st, memSess(), memIDs(), memAcc(), memTok(),
		trustDec{dec: ports.TrustDecision{Allow: true}},
		fakeAudit{}, 5*time.Minute, 7*24*time.Hour, "pepper",
	)

	out, err := u.GoogleCallback(context.Background(), GoogleCallbackInput{
		Code: "c", State: "state", RedirectURL: "http://cb",
		Client: ClientInfo{DeviceID: "d", UserAgentHash: "ua"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.TrustLevel != "aal2" || out.StepUpRequired {
		t.Fatalf("want aal2 + no stepup, got trust=%s stepup=%v", out.TrustLevel, out.StepUpRequired)
	}
}

func TestGoogleCallback_StepUp_DoesNotElevate_WhenTrustRequiresStepUp(t *testing.T) {
	st := memState()
	_ = st.Put(context.Background(), "state",
		ports.AuthState{Nonce: "n", CodeVerifier: "v", Purpose: "stepup", CreatedAt: time.Now()},
		time.Minute,
	)

	u, _ := NewGoogleFlow(oidcOK{}, st, memSess(), memIDs(), memAcc(), memTok(),
		trustDec{dec: ports.TrustDecision{Allow: true, RequireStepUp: true}},
		fakeAudit{}, 5*time.Minute, 7*24*time.Hour, "pepper",
	)

	out, err := u.GoogleCallback(context.Background(), GoogleCallbackInput{
		Code: "c", State: "state", RedirectURL: "http://cb",
		Client: ClientInfo{DeviceID: "d", UserAgentHash: "ua"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.TrustLevel != "aal1" || !out.StepUpRequired {
		t.Fatalf("want aal1 + stepup, got trust=%s stepup=%v", out.TrustLevel, out.StepUpRequired)
	}
}
