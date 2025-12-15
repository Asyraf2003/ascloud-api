package presenter

type Meta struct {
	RequestID string `json:"request_id,omitempty"`
}

type SuccessEnvelope struct {
	Data any   `json:"data"`
	Meta *Meta `json:"meta,omitempty"`
}

type AuthEnvelope struct {
	Auth any   `json:"auth"`
	Meta *Meta `json:"meta,omitempty"`
}

type BillingEnvelope struct {
	Billing any   `json:"billing"`
	Meta    *Meta `json:"meta,omitempty"`
}

type ErrorEnvelope struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}
