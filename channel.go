package lnurl

import (
	"encoding/json"
	"errors"
	"net/url"
)

type LNURLChannelResponse struct {
	LNURLResponse
	Tag         string   `json:"tag"`
	K1          string   `json:"k1"`
	Callback    string   `json:"callback"`
	CallbackURL *url.URL `json:"-"`
	URI         string   `json:"uri"`
}

func (_ LNURLChannelResponse) LNURLKind() string { return "lnurl-channel" }

func HandleChannel(raw []byte) (LNURLParams, error) {
	var params LNURLChannelResponse
	err := json.Unmarshal(raw, &params)
	if err != nil {
		return nil, err
	}

	callbackURL, err := url.Parse(params.Callback)
	if err != nil {
		return nil, errors.New("callback is not a valid URL")
	}
	params.CallbackURL = callbackURL

	return params, nil
}
