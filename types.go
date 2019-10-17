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

type LNURLParams interface {
	LNURLKind() string
}

type LNURLChannelResponse struct {
	LNURLResponse
	Tag      string `json:"tag"`
	K1       string `json:"k1"`
	Callback string `json:"callback"`
	URI      string `json:"uri"`
}

func (_ LNURLChannelResponse) LNURLKind() string { return "lnurl-channel" }

type LNURLWithdrawResponse struct {
	LNURLResponse
	Tag                string `json:"tag"`
	K1                 string `json:"k1"`
	Callback           string `json:"callback"`
	MaxWithdrawable    int64  `json:"maxWithdrawable"`
	MinWithdrawable    int64  `json:"minWithdrawable"`
	DefaultDescription string `json:"defaultDescription"`
}

func (_ LNURLWithdrawResponse) LNURLKind() string { return "lnurl-withdraw" }

type LNURLPayResponse struct {
	LNURLResponse
	Tag         string      `json:"tag"`
	MaxSendable int64       `json:"maxWithdrawable"`
	MinSendable int64       `json:"minWithdrawable"`
	Routes      interface{} `json:"routes"`
	PR          string      `json:"pr"`
	Metadata    string      `json:"metadata"`
}

func (_ LNURLPayResponse) LNURLKind() string { return "lnurl-pay" }

type LNURLAuthParams struct {
	Tag      string
	K1       string
	Callback string
	Host     string
}

func (_ LNURLAuthParams) LNURLKind() string { return "lnurl-auth" }
