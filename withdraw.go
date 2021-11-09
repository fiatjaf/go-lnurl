package lnurl

import (
	"encoding/json"
	"errors"
	"net/url"
	"strconv"
)

type LNURLWithdrawResponse struct {
	LNURLResponse
	Tag                string   `json:"tag"`
	K1                 string   `json:"k1"`
	Callback           string   `json:"callback"`
	CallbackURL        *url.URL `json:"-"`
	MaxWithdrawable    int64    `json:"maxWithdrawable"`
	MinWithdrawable    int64    `json:"minWithdrawable"`
	DefaultDescription string   `json:"defaultDescription"`
	BalanceCheck       string   `json:"balanceCheck,omitempty"`
	PayLink            string   `json:"payLink,omitempty"`
}

func (_ LNURLWithdrawResponse) LNURLKind() string { return "lnurl-withdraw" }

func HandleWithdraw(raw []byte) (LNURLParams, error) {
	var params LNURLWithdrawResponse
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

func HandleFastWithdraw(query url.Values) (LNURLParams, bool) {
	callback := query.Get("callback")
	if callback == "" {
		return nil, false
	}
	callbackURL, err := url.Parse(callback)
	if err != nil {
		return nil, false
	}
	maxWithdrawable, err := strconv.ParseInt(query.Get("maxWithdrawable"), 10, 64)
	if err != nil {
		return nil, false
	}
	minWithdrawable, err := strconv.ParseInt(query.Get("minWithdrawable"), 10, 64)
	if err != nil {
		return nil, false
	}
	balanceCheck := query.Get("balanceCheck")
	payLink := query.Get("payLink")

	return LNURLWithdrawResponse{
		Tag:                "withdrawRequest",
		K1:                 query.Get("k1"),
		Callback:           callback,
		CallbackURL:        callbackURL,
		MaxWithdrawable:    maxWithdrawable,
		MinWithdrawable:    minWithdrawable,
		DefaultDescription: query.Get("defaultDescription"),
		BalanceCheck:       balanceCheck,
		PayLink:            payLink,
	}, true
}
