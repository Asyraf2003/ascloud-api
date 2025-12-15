package usecase

import "time"

type RefreshInput struct {
	RefreshToken string
	Client       ClientInfo
}

type RefreshOutput struct {
	AccountID, SessionID string
	AccessToken          string
	AccessExp            time.Time
	RefreshToken         string
	RefreshExp           time.Time
	CSRFToken            string
	TrustLevel           string
	StepUpRequired       bool
}
