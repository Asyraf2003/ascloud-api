package ports

import "context"

type TrustSignals struct {
	AccountID string
	Purpose   string // login|stepup|refresh
	IP        string
	UserAgent string
}

type TrustDecision struct {
	Allow         bool
	RequireStepUp bool
	Reason        string
}

type TrustEvaluator interface {
	Evaluate(ctx context.Context, s TrustSignals) (TrustDecision, error)
}
