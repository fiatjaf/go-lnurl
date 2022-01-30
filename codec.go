package lnurl

import (
	"errors"
	"net/url"
	"strings"
)

// LNURLDecode takes a lowercase bech32-encoded lnurl string and returns a plain-text https URL.
func LNURLDecode(lnurl string) (string, error) {
	switch {
	case strings.HasPrefix(lnurl, "lnurl1"):
		// bech32
		tag, data, err := Decode(lnurl)
		if err != nil {
			return "", err
		}

		if tag != "lnurl" {
			return "", errors.New("tag is not 'lnurl', but '" + tag + "'")
		}

		converted, err := ConvertBits(data, 5, 8, false)
		if err != nil {
			return "", err
		}

		return string(converted), nil
	case strings.HasPrefix(lnurl, "lnurlp://"),
		strings.HasPrefix(lnurl, "lnurlw://"),
		strings.HasPrefix(lnurl, "lnurlc://"),
		strings.HasPrefix(lnurl, "keyauth://"),
		strings.HasPrefix(lnurl, "https://"):

		u := "https://" + strings.SplitN(lnurl, "://", 2)[1]
		if parsed, err := url.Parse(u); err == nil &&
			strings.HasSuffix(parsed.Host, ".onion") {
			u = "https://" + strings.SplitN(lnurl, "://", 2)[1]
		}

		return u, nil
	}

	return "", errors.New("unrecognized lnurl format: " + lnurl)
}

// LNURLEncode takes a plain-text https URL and returns a bech32-encoded uppercased lnurl string.
func LNURLEncode(actualurl string) (lnurl string, err error) {
	asbytes := []byte(actualurl)
	converted, err := ConvertBits(asbytes, 8, 5, true)
	if err != nil {
		return
	}

	lnurl, err = Encode("lnurl", converted)
	return strings.ToUpper(lnurl), err
}
