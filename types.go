package lnurl

import (
	"encoding/base64"
	"net/url"
)

// The base response for all lnurl calls.
type LNURLResponse struct {
	Status string `json:"status,omitempty"`
	Reason string `json:"reason,omitempty"`
}

func OkResponse() LNURLResponse {
	return LNURLResponse{Status: "OK"}
}

func ErrorResponse(reason string) LNURLResponse {
	return LNURLResponse{Status: "ERROR", Reason: reason}
}

func Action(text string, url string) *SuccessAction {
	if url == "" {
		return &SuccessAction{
			Tag:     "message",
			Message: text,
		}
	}

	return &SuccessAction{
		Tag:         "url",
		Description: text,
		URL:         url,
	}
}

func NoAction() *SuccessAction {
	return &SuccessAction{Tag: "noop"}
}

func AESAction(description string, preimage []byte, content string) (*SuccessAction, error) {
	plaintext := []byte(content)

	ciphertext, iv, err := AESCipher(preimage, plaintext)
	if err != nil {
		return nil, err
	}

	return &SuccessAction{
		Tag:         "aes",
		Description: description,
		Ciphertext:  base64.StdEncoding.EncodeToString(ciphertext),
		IV:          base64.StdEncoding.EncodeToString(iv),
	}, nil
}

type LNURLParams interface {
	LNURLKind() string
}

type LNURLChannelResponse struct {
	LNURLResponse
	Tag         string   `json:"tag"`
	K1          string   `json:"k1"`
	Callback    string   `json:"callback"`
	CallbackURL *url.URL `json:"-"`
	URI         string   `json:"uri"`
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
	SuccessAction *SuccessAction `json:"successAction,omitempty"`
	Routes        [][]RouteInfo  `json:"routes"`
	PR            string         `json:"pr"`
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
	Ciphertext  string `json:"ciphertext,omitempty"`
	IV          string `json:"iv,omitempty"`
}

func (sa *SuccessAction) Decipher(preimage []byte) (content string, err error) {
	ciphertext, err := base64.StdEncoding.DecodeString(sa.Ciphertext)
	if err != nil {
		return
	}

	iv, err := base64.StdEncoding.DecodeString(sa.IV)
	if err != nil {
		return
	}

	plaintext, err := AESDecipher(preimage, ciphertext, iv)
	if err != nil {
		return
	}

	return string(plaintext), nil
}

func (_ LNURLPayResponse1) LNURLKind() string { return "lnurl-pay" }

type LNURLAuthParams struct {
	Tag      string
	K1       string
	Callback string
	Host     string
}

func (_ LNURLAuthParams) LNURLKind() string { return "lnurl-auth" }
