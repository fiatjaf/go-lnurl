package lnurl

import (
	"crypto/rand"
	"math/big"
	"regexp"
	"strings"
)

var lnurlregex = regexp.MustCompile(`.*?((lnurl)([0-9]{1,}[a-z0-9]+){1})`)

// FindLNURLInText uses a Regular Expression to find a bech32-encoded lnurl string in a blob of text.
func FindLNURLInText(text string) (lnurl string, ok bool) {
	text = strings.ToLower(text)
	results := lnurlregex.FindStringSubmatch(text)

	if len(results) == 0 {
		return
	}

	return results[1], true
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
