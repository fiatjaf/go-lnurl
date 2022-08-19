package lnurl

import (
	"errors"
	"fmt"
	"golang.org/x/net/idna"
	"golang.org/x/net/publicsuffix"
	"net"
	"net/url"
	"strings"
)

var lud17ValidSchemes = map[string]struct{}{"lnurla": {}, "lnurlp": {}, "lnurlw": {}, "lnurlc": {}, "keyauth": {}}

// LNURLDecode takes a bech32-encoded lnurl string and returns a plain-text https URL.
func LNURLDecode(code string) (string, error) {
	code = strings.ToLower(code)

	switch {
	case strings.HasPrefix(code, "lnurl1"):
		// bech32
		tag, data, err := Decode(code)
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
	case strings.HasPrefix(code, "lnurlp://"),
		strings.HasPrefix(code, "lnurlw://"),
		strings.HasPrefix(code, "lnurlc://"),
		strings.HasPrefix(code, "keyauth://"),
		strings.HasPrefix(code, "https://"):

		u := "https://" + strings.SplitN(code, "://", 2)[1]
		if parsed, err := url.Parse(u); err == nil &&
			strings.HasSuffix(parsed.Host, ".onion") {
			u = "https://" + strings.SplitN(code, "://", 2)[1]
		}

		return u, nil
	}

	return "", errors.New("unrecognized lnurl format: " + code)
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

// LURL embeds net/url and adds extra fields ontop
type lnUrl struct {
	subdomain, domain, tld, port, publicSuffix string
	icann, isDomain, isIp                      bool
	*url.URL
}

func (u lnUrl) String() string {
	decodedValue, err := url.QueryUnescape(u.URL.String())
	if err != nil {
		return ""
	}
	return decodedValue
}

//parse mirrors net/url.Parse except instead it returns
//a tld.URL, which contains extra fields.
func parse(s string) (*lnUrl, error) {
	s = addDefaultScheme(s)
	parsedUrl, err := url.Parse(s)
	if err != nil {
		return nil, err
	}
	if parsedUrl.Host == "" {
		return &lnUrl{URL: parsedUrl}, nil
	}

	dom, port := domainPort(parsedUrl.Host)
	//etld+1
	etld1, err := publicsuffix.EffectiveTLDPlusOne(dom)
	if err != nil {
		return nil, err
	}
	//convert to domain name, and tld
	i := strings.Index(etld1, ".")
	domName := etld1[0:i]
	tld := etld1[i+1:]
	//and subdomain
	sub := ""
	if rest := strings.TrimSuffix(dom, "."+etld1); rest != dom {
		sub = rest
	}
	s, err = idna.New(idna.ValidateForRegistration()).ToASCII(dom)
	if err != nil {
		return nil, err
	}
	psuf, icann := publicsuffix.PublicSuffix(tld)
	return &lnUrl{
		subdomain:    sub,
		domain:       domName,
		tld:          tld,
		port:         port,
		URL:          parsedUrl,
		publicSuffix: psuf,
		icann:        icann,
		isDomain:     IsDomainName(s),
		isIp:         net.ParseIP(dom) != nil,
	}, nil
}

// adds default scheme //, if nothing is defined at all
func addDefaultScheme(s string) string {
	if strings.Index(s, "//") == -1 {
		return fmt.Sprintf("//%s", s)
	}
	return s
}

// domainPort splits domain.com:8080
func domainPort(host string) (string, string) {
	for i := len(host) - 1; i >= 0; i-- {
		if host[i] == ':' {
			return host[:i], host[i+1:]
		} else if host[i] < '0' || host[i] > '9' {
			return host, ""
		}
	}
	//will only land here if the string is all digits,
	//net/url should prevent that from happening
	return host, ""
}

// LNURLDecode takes a string and returns a valid lnurl, if possible.
// code can be
func LNURLDecodeStrict(code string) (string, error) {
	code = strings.ToLower(code)
	switch {
	case strings.HasPrefix(code, "lnurl1"):
		// bech32
		tag, data, err := Decode(code)
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
		u, err := parse(string(converted))
		if err != nil {
			return string(converted), err
		}
		if u.isIp {
			if u.Scheme != "https" {
				err := fmt.Errorf("invalid scheme: %s", string(converted))
				u.Scheme = "https"
				return u.String(), err
			}
			return u.String(), nil
		}
		if !u.isDomain {
			return string(converted), fmt.Errorf("invalid domain: %s", string(converted))
		}
		if setScheme(u) {
			return u.String(), fmt.Errorf("invalid scheme: %s", u.Scheme)
		}
		return u.String(), nil
	case strings.HasPrefix(code, "https://"):
		return code, nil
	default:
		u, err := parse(code)
		if err != nil {
			return "", err
		}
		scheme := u.Scheme
		lud17 := validLud17(scheme)
		if setScheme(u) {
			if !lud17 {
				return u.String(), fmt.Errorf("invalid scheme: %s", scheme)
			}
		}
		return u.String(), nil
	}
}

// setHttpsScheme will parse string url to url.Url.
// if no scheme was found,
func setScheme(u *lnUrl) (updated bool) {
	if u.tld == "onion" {
		if u.Scheme != "http" {
			u.Scheme = "http"
			updated = true
		}
	} else {
		if u.Scheme != "https" {
			u.Scheme = "https"
			updated = true
		}
	}
	return
}

// validLud17 will return true, if scheme is valid for lud17
func validLud17(schema string) bool {
	_, ok := lud17ValidSchemes[schema]
	return ok
}

// LNURLEncodeStrict will encode the actualurl to lnurl.
// based on the input url, it will determine whether bech32 encoding / url manipulation is necessary
func LNURLEncodeStrict(actualurl string) (string, error) {
	lnurl, err := parse(actualurl)
	if err != nil {
		enc, encErr := encode(actualurl)
		if encErr != nil {
			return "", encErr
		}
		return enc, fmt.Errorf("invalid url: %s", actualurl)
	}
	if validLud17(lnurl.Scheme) {
		return lnurl.String(), nil
	}
	if lnurl.isIp {
		// actualurl is an ip. just change scheme
		lnurl.Scheme = "https"
		return encode(lnurl.String())
	}
	if !lnurl.isDomain {
		enc, encErr := encode(actualurl)
		if encErr != nil {
			return "", encErr
		}
		return enc, fmt.Errorf("invalid domain: %s", lnurl.tld)
	}
	updated := false
	if lnurl.tld != "onion" {
		// check tld
		if !lnurl.icann {
			enc, encErr := encode(actualurl)
			if encErr != nil {
				return "", encErr
			}
			return enc, fmt.Errorf("invalid tld: %s", lnurl.tld)
		}
		if lnurl.Scheme != "https" {
			lnurl.Scheme = "https"
			updated = true
		}

	} else {
		if lnurl.Scheme != "http" {
			lnurl.Scheme = "http"
			updated = true
		}
	}
	enc, err := encode(lnurl.String())
	if err != nil {
		return enc, err
	}
	if updated {
		return enc, fmt.Errorf("invalid protocol schema: %s", lnurl.Scheme)
	}
	return enc, err
}

func encode(s string) (string, error) {
	asbytes := []byte(s)
	converted, err := ConvertBits(asbytes, 8, 5, true)
	if err != nil {
		return s, err
	}

	lnurl, err := Encode("lnurl", converted)
	return strings.ToUpper(lnurl), err
}

// IsDomainName (from net package) checks if a string is a presentation-format domain name
// (currently restricted to hostname-compatible "preferred name" LDH labels and
// SRV-like "underscore labels"; see golang.org/issue/12421).
func IsDomainName(s string) bool {
	// The root domain name is valid. See golang.org/issue/45715.
	if s == "." {
		return true
	}

	// See RFC 1035, RFC 3696.
	// Presentation format has dots before every label except the first, and the
	// terminal empty label is optional here because we assume fully-qualified
	// (absolute) input. We must therefore reserve space for the first and last
	// labels' length octets in wire format, where they are necessary and the
	// maximum total length is 255.
	// So our _effective_ maximum is 253, but 254 is not rejected if the last
	// character is a dot.
	l := len(s)
	if l == 0 || l > 254 || l == 254 && s[l-1] != '.' {
		return false
	}

	last := byte('.')
	nonNumeric := false // true once we've seen a letter or hyphen
	partlen := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		default:
			return false
		case 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || c == '_':
			nonNumeric = true
			partlen++
		case '0' <= c && c <= '9':
			// fine
			partlen++
		case c == '-':
			// Byte before dash cannot be dot.
			if last == '.' {
				return false
			}
			partlen++
			nonNumeric = true
		case c == '.':
			// Byte before dot cannot be dot, dash.
			if last == '.' || last == '-' {
				return false
			}
			if partlen > 63 || partlen == 0 {
				return false
			}
			partlen = 0
		}
		last = c
	}
	if last == '-' || partlen > 63 {
		return false
	}

	return nonNumeric
}
