package lnurl

import (
	"encoding/base64"
	"errors"
	"net/url"

	"github.com/tidwall/gjson"
)

type LNURLAllowanceResponse struct {
	LNURLResponse
	Tag                        string   `json:"tag"`
	K1                         string   `json:"k1"`
	RecommendedAllowanceAmount int64    `json:"RecommendedAllowanceAmount"`
	Image                      string   `json:"image"`
	Description                string   `json:"description"`
	Socket                     string   `json:"socket"`
	SocketURL                  *url.URL `json:"-"`
}

func (_ LNURLAllowanceResponse) LNURLKind() string { return "lnurl-allowance" }

func HandleAllowance(j gjson.Result) (LNURLParams, error) {
	socket := j.Get("socket").String()
	socketURL, err := url.Parse(socket)
	if err != nil || socketURL.Scheme != "wss" {
		return nil, errors.New("socket " + socket + " is not a valid URL")
	}

	return LNURLAllowanceResponse{
		Tag:                        "allowanceRequest",
		K1:                         j.Get("k1").String(),
		Socket:                     socket,
		SocketURL:                  socketURL,
		Image:                      j.Get("image").String(),
		Description:                j.Get("description").String(),
		RecommendedAllowanceAmount: j.Get("recommendedAllowanceAmount").Int(),
	}, nil
}

// ImageBytes returns image as bytes, decoded from base64 if an image exists
// or nil if not.
func (resp LNURLAllowanceResponse) ImageBytes() []byte {
	if decoded, err := base64.StdEncoding.DecodeString(resp.Image); err == nil {
		return decoded
	}
	return nil
}

// messages sent over the websocket

type AllowanceRequest struct {
	K1      string `json:"k1"`
	Balance int64  `json:"balance"`
	Type    string `json:"allowanceRequest"`
}

type AllowanceSuccess struct {
	LNURLResponse

	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
	Type      string `json:"type"`
}

type AllowanceMessage struct {
	Type string `json:"type"`

	Amount  int64  `json:"amount"`  // for "invoiceRequest"
	Invoice string `json:"invoice"` // for "paymentRequest" and "invoiceSuccess"
}
