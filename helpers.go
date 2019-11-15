package lnurl

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"math/big"
	"regexp"
	"strings"

	"github.com/btcsuite/btcd/btcec"
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
