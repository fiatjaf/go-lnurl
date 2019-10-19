package lnurl

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/btcsuite/btcd/btcec"
	"github.com/tidwall/gjson"
)

var lnurlregex = regexp.MustCompile(`,*?((lnurl)([0-9]{1,}[a-z0-9]+){1})`)

// FindLNURLInText uses a Regular Expression to find a bech32-encoded lnurl string in a blob of text.
func FindLNURLInText(text string) (lnurl string, ok bool) {
	text = strings.ToLower(text)
	results := lnurlregex.FindStringSubmatch(text)

	if len(results) == 0 {
		return
	}

	return results[1], true
}

// VerifySignature takes the hex-encoded parameters passed to an lnurl-login endpoint and verifies
// the signature against the key and challenge.
func VerifySignature(k1, sig, key string) (ok bool, err error) {
	bk1, err1 := hex.DecodeString(k1)
	bsig, err2 := hex.DecodeString(sig)
	bkey, err3 := hex.DecodeString(key)
	if err1 != nil || err2 != nil || err3 != nil {
		return false, errors.New("Failed to decode hex.")
	}

	pubkey, err := btcec.ParsePubKey(bkey, btcec.S256())
	if err != nil {
		return false, errors.New("Failed to parse pubkey: " + err.Error())
	}

	signature, err := btcec.ParseSignature(bsig, btcec.S256())
	if err != nil {
		return false, errors.New("Failed to parse signature: " + err.Error())
	}

	return signature.Verify(bk1, pubkey), nil
}

// HandleLNURL takes a bech32-encoded lnurl and either gets its parameters from the query-
// string or calls the URL to get the parameters.
// Returns a different struct for each of the lnurl subprotocols, the .LNURLKind() method of
// which should be checked next to see how the wallet is going to proceed.
func HandleLNURL(rawlnurl string) (LNURLParams, error) {
	lnurl, ok := FindLNURLInText(rawlnurl)
	if !ok {
		return nil, errors.New("invalid bech32-encoded lnurl: " + rawlnurl)
	}

	rawurl, err := LNURLDecode(lnurl)
	if err != nil {
		return nil, err
	}

	parsed, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}

	query := parsed.Query()

	if query.Get("tag") == "login" {
		k1 := query.Get("k1")
		if _, err := hex.DecodeString(k1); err != nil || len(k1) != 64 {
			return nil, errors.New("k1 is not a valid 32-byte hex-encoded string.")
		}

		return LNURLAuthParams{
			Tag:      "login",
			K1:       k1,
			Callback: rawurl,
			Host:     parsed.Host,
		}, nil
	} else {
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
			return nil, errors.New(j.Get("reason").String())
		}

		switch j.Get("tag").String() {
		case "channelRequest":
			k1 := j.Get("k1").String()
			if k1 == "" {
				return nil, errors.New("k1 is blank")
			}
			callback := j.Get("callback").String()
			if urlp, err := url.Parse(callback); err != nil || urlp.Scheme != "https" {
				return nil, errors.New("callback is not a valid HTTPS URL")
			}

			return LNURLChannelResponse{
				Tag:      "channelRequest",
				K1:       k1,
				Callback: callback,
				URI:      j.Get("uri").String(),
			}, nil
		case "withdrawRequest":
			k1 := j.Get("k1").String()
			if k1 == "" {
				return nil, errors.New("k1 is blank")
			}
			callback := j.Get("callback").String()
			if urlp, err := url.Parse(callback); err != nil || urlp.Scheme != "https" {
				return nil, errors.New("callback is not a valid HTTPS URL")
			}

			return LNURLWithdrawResponse{
				Tag:                "withdrawRequest",
				K1:                 k1,
				Callback:           callback,
				MaxWithdrawable:    j.Get("maxWithdrawable").Int(),
				MinWithdrawable:    j.Get("minWithdrawable").Int(),
				DefaultDescription: j.Get("defaultDescription").String(),
			}, nil
		case "payRequest":
			return LNURLPayResponse{
				Tag:         "payRequest",
				Routes:      j.Get("routes").Value(),
				PR:          j.Get("pr").String(),
				Metadata:    j.Get("metadata").String(),
				MaxSendable: j.Get("maxSendable").Int(),
				MinSendable: j.Get("minSendable").Int(),
			}, nil
		default:
			return nil, errors.New("unknown response tag " + j.Get("tag").String())
		}
	}
}

// RandomK1 returns a 32-byte random hex-encoded string for usage as k1 in lnurl-auth and anywhere else.
func RandomK1() string {
	hex := []rune("0123456789abcdef")
	b := make([]rune, 64)
	for i := range b {
		r, err := rand.Int(rand.Reader, big.NewInt(16))
		if err != nil {
			return ""
		}
		b[i] = hex[r.Int64()]
	}
	return string(b)
}
