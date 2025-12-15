package domain

type Provider string

const (
	ProviderGoogle Provider = "google"
)

type AssuranceLevel string

const (
	AAL1 AssuranceLevel = "aal1"
	AAL2 AssuranceLevel = "aal2"
)

type AuthPurpose string

const (
	PurposeLogin  AuthPurpose = "login"
	PurposeStepUp AuthPurpose = "stepup"
)
