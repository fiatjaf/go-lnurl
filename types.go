package lnurl

// The base response for all lnurl calls.
type LNURLResponse struct {
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
}

func OkResponse() LNURLResponse {
	return LNURLResponse{Status: "OK"}
}

func ErrorResponse(reason string) LNURLResponse {
	return LNURLResponse{Status: "ERROR", Reason: reason}
}

// The response a service must return when a wallet scans an lnurl-withdraw QR code.
// Must include Status: "OK".
type LNURLWithdrawResponse struct {
	Callback           string `json:"callback"`
	K1                 string `json:"k1"`
	MaxWithdrawable    int64  `json:"maxWithdrawable"`
	MinWithdrawable    int64  `json:"minWithdrawable"`
	DefaultDescription string `json:"defaultDescription"`
	Tag                string `json:"tag"`
	LNURLResponse
}
