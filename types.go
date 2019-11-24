package lnurl

import (
	"net/url"
)

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

func Action(text string, url string) SuccessAction {
	if url == "" {
		return SuccessAction{
			Tag:     "message",
			Message: text,
		}
	}

	return SuccessAction{
		Tag:         "message",
		Description: text,
		URL:         url,
	}
}

func NoAction() SuccessAction {
	return SuccessAction{Tag: "noop"}
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
	Tag                string   `json:"tag"`
	K1                 string   `json:"k1"`
	Callback           string   `json:"callback"`
	CallbackURL        *url.URL `json:"-"`
	MaxWithdrawable    int64    `json:"maxWithdrawable"`
	MinWithdrawable    int64    `json:"minWithdrawable"`
	DefaultDescription string   `json:"defaultDescription"`
}

func (_ LNURLWithdrawResponse) LNURLKind() string { return "lnurl-withdraw" }

type LNURLPayResponse1 struct {
	LNURLResponse
	Callback        string     `json:"callback"`
	CallbackURL     *url.URL   `json:"-"`
	Tag             string     `json:"tag"`
	MaxSendable     int64      `json:"maxSendable"`
	MinSendable     int64      `json:"minSendable"`
	EncodedMetadata string     `json:"metadata"`
	Metadata        [][]string `json:"-"`
}

type LNURLPayResponse2 struct {
	LNURLResponse
	SuccessAction SuccessAction `json:"successAction"`
	Routes        [][]RouteInfo `json:"routes"`
	PR            string        `json:"pr"`
}

type RouteInfo struct {
	NodeId        string `json:"nodeId"`
	ChannelUpdate string `json:"channelUpdate"`
}

type SuccessAction struct {
	Tag         string `json:"tag"`
	Description string `json:"description,omitempty"`
	URL         string `json:"url,omitempty"`
	Message     string `json:"message,omitempty"`
}

func (_ LNURLPayResponse1) LNURLKind() string { return "lnurl-pay" }

type LNURLAuthParams struct {
	Tag      string
	K1       string
	Callback string
	Host     string
}

func (_ LNURLAuthParams) LNURLKind() string { return "lnurl-auth" }
