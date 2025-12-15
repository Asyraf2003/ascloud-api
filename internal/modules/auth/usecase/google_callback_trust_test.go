package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"example.com/your-api/internal/modules/auth/ports"
)

type oidcDummy struct{}

func (o oidcDummy) AuthCodeURL(p ports.OIDCAuthURLParams) string { return "u" }
func (o oidcDummy) ExchangeAndVerify(ctx context.Context, code, v, r, n string) (ports.OIDCClaims, error) {
	return ports.OIDCClaims{}, nil
}

type trustOK struct{ d ports.TrustDecision }

func (t trustOK) Evaluate(ctx context.Context, s ports.TrustSignals) (ports.TrustDecision, error) {
	return t.d, nil
}

type trustErr struct{}

func (t trustErr) Evaluate(ctx context.Context, s ports.TrustSignals) (ports.TrustDecision, error) {
	return ports.TrustDecision{}, errors.New("boom")
}

func newFlow(t ports.TrustEvaluator) *GoogleFlow {
	u, err := NewGoogleFlow(oidcDummy{}, memState(), memSess(), memIDs(), memAcc(), memTok(), t, fakeAudit{},
		time.Minute, time.Hour, "pepper")
	if err != nil {
		panic(err)
	}
	return u
}

func TestDecideTrustAndAAL_StepUp_OK_GetsAAL2(t *testing.T) {
	u := newFlow(trustOK{d: ports.TrustDecision{Allow: true}})
	lvl, step, err := decideTrustAndAAL(u, context.Background(), "acc", "stepup", ClientInfo{UserAgentHash: "ua"})
	if err != nil || lvl != "aal2" || step {
		t.Fatalf("lvl=%s step=%v err=%v", lvl, step, err)
	}
}

func TestDecideTrustAndAAL_StepUp_TrustError_NoAAL2(t *testing.T) {
	u := newFlow(trustErr{})
	lvl, step, err := decideTrustAndAAL(u, context.Background(), "acc", "stepup", ClientInfo{UserAgentHash: "ua"})
	if err != nil || lvl != "aal1" || !step {
		t.Fatalf("lvl=%s step=%v err=%v", lvl, step, err)
	}
}
