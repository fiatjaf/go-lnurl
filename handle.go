package lnurl

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/tidwall/gjson"
)

// HandleLNURL takes a bech32-encoded lnurl and either gets its parameters from the query-
// string or calls the URL to get the parameters.
// Returns a different struct for each of the lnurl subprotocols, the .LNURLKind() method of
// which should be checked next to see how the wallet is going to proceed.
func HandleLNURL(rawlnurl string) (LNURLParams, error) {
	var err error
	var rawurl string

	if strings.HasPrefix(rawlnurl, "https:") {
		rawurl = rawlnurl
	} else {
		lnurl, ok := FindLNURLInText(rawlnurl)
		if !ok {
			return nil, errors.New("invalid bech32-encoded lnurl: " + rawlnurl)
		}
		rawurl, err = LNURLDecode(lnurl)
		if err != nil {
			return nil, err
		}
	}

	parsed, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}

	query := parsed.Query()

	switch query.Get("tag") {
	case "login":
		return HandleAuth(rawurl, parsed, query)
	case "withdrawRequest":
		if value, ok := HandleFastWithdraw(query); ok {
			return value, nil
		}
	}

	resp, err := http.Get(rawurl)
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	j := gjson.ParseBytes(b)
	if j.Get("status").String() == "ERROR" {
		return nil, LNURLErrorResponse{
			URL:    parsed,
			Reason: j.Get("reason").String(),
			Status: "ERROR",
		}
	}

	switch j.Get("tag").String() {
	case "withdrawRequest":
		return HandleWithdraw(j)
	case "payRequest":
		return HandlePay(j)
	case "channelRequest":
		return HandleChannel(j)
	default:
		return nil, errors.New("unknown response tag " + j.String())
	}
}
