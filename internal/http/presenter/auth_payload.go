package presenter

type AuthTokens struct {
	AccessToken     string `json:"access_token"`
	AccessExpiresAt int64  `json:"access_expires_at"`
	TrustLevel      string `json:"trust_level"`
	StepUpRequired  bool   `json:"step_up_required"`
}
