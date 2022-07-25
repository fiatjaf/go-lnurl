package lnurl

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"strings"
)

var (
	lud01regex = regexp.MustCompile(`\b((lnurl)([0-9]{1,}[a-z0-9]+){1})\b`)
	lud17regex = regexp.MustCompile(`^ *((lnurlp|lnurlw|lnurlc|keyauth):\/\/\S+) *$`)
)

// FindLNURLInText uses a Regular Expression to find a bech32-encoded lnurl string in a blob of text.
func FindLNURLInText(text string) (lnurl string, ok bool) {
	results := lud01regex.FindStringSubmatch(strings.ToLower(text))
	if len(results) == 0 {
		results = lud17regex.FindStringSubmatch(text)
		if len(results) == 3 {
			return results[1], true
		}
		return
	}
	return results[1], true
}

// RandomK1 returns a 32-byte random hex-encoded string for usage as k1 in lnurl-auth and anywhere else.
func RandomK1() string {
	random := make([]byte, 32)
	rand.Read(random)
	return hex.EncodeToString(random)
}

// ParseInternetIdentifier extracts name and domain from an email-like string like username@example.com
func ParseInternetIdentifier(text string) (name, domain string, ok bool) {
	nameAndDomain := strings.Split(text, "@")
	if len(nameAndDomain) != 2 {
		return
	}

	name = nameAndDomain[0]
	domain = nameAndDomain[1]
	if len(name) == 0 || len(domain) == 0 {
		return
	}

	if strings.Index(domain, ".") == -1 {
		return
	}

	ok = true
	return
}
