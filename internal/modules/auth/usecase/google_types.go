package usecase

import "time"

type ClientInfo struct {
	DeviceID      string
	UserAgentHash string
	IPPrefix      *string
}

type GoogleStartInput struct {
	Purpose     string // login|stepup
	RedirectURL string
}

type GoogleStartOutput struct {
	RedirectTo string
	State      string
}

type GoogleCallbackInput struct {
	Code        string
	State       string
	RedirectURL string
	Client      ClientInfo
}

type GoogleCallbackOutput struct {
	AccountID string
	SessionID string

	AccessToken string
	AccessExp   time.Time

	RefreshToken string
	RefreshExp   time.Time

	CSRFToken string

	TrustLevel     string
	StepUpRequired bool
}
