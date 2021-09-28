package lnurl

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/tidwall/gjson"
)

var (
	f     bool  = false
	t     bool  = true
	FALSE *bool = &f
	TRUE  *bool = &t
)

func Action(text string, url string) *SuccessAction {
	if url == "" {
		return &SuccessAction{
			Tag:     "message",
			Message: text,
		}
	}

	if text == "" {
		text = " "
	}
	return &SuccessAction{
		Tag:         "url",
		Description: text,
		URL:         url,
	}
}

func AESAction(description string, preimage []byte, content string) (*SuccessAction, error) {
	plaintext := []byte(content)

	ciphertext, iv, err := AESCipher(preimage, plaintext)
	if err != nil {
		return nil, err
	}

	return &SuccessAction{
		Tag:         "aes",
		Description: description,
		Ciphertext:  base64.StdEncoding.EncodeToString(ciphertext),
		IV:          base64.StdEncoding.EncodeToString(iv),
	}, nil
}

type LNURLPayResponse1 struct {
	LNURLResponse
	Callback       string   `json:"callback"`
	CallbackURL    *url.URL `json:"-"`
	Tag            string   `json:"tag"`
	MaxSendable    int64    `json:"maxSendable"`
	MinSendable    int64    `json:"minSendable"`
	Metadata       Metadata `json:"metadata"`
	CommentAllowed int64    `json:"commentAllowed"`
}

type LNURLPayResponse2 struct {
	LNURLResponse
	SuccessAction *SuccessAction `json:"successAction"`
	Routes        [][]RouteInfo  `json:"routes"`
	PR            string         `json:"pr"`
	Disposable    *bool          `json:"disposable,omitempty"`
}

type RouteInfo struct {
	NodeId        string `json:"nodeId"`
	ChannelUpdate string `json:"channelUpdate"`
}

type SuccessAction struct {
	Tag         string `json:"tag"`
	Description string `json:"description,omitempty"`
	URL         string `json:"url,omitempty"`
	Message     string `json:"message,omitempty"`
	Ciphertext  string `json:"ciphertext,omitempty"`
	IV          string `json:"iv,omitempty"`
}

func (sa *SuccessAction) Decipher(preimage []byte) (content string, err error) {
	ciphertext, err := base64.StdEncoding.DecodeString(sa.Ciphertext)
	if err != nil {
		return
	}

	iv, err := base64.StdEncoding.DecodeString(sa.IV)
	if err != nil {
		return
	}

	plaintext, err := AESDecipher(preimage, ciphertext, iv)
	if err != nil {
		return
	}

	return string(plaintext), nil
}

func (_ LNURLPayResponse1) LNURLKind() string { return "lnurl-pay" }

func HandlePay(j gjson.Result) (LNURLParams, error) {
	strmetadata := j.Get("metadata").String()
	var metadata Metadata
	err := json.Unmarshal([]byte(strmetadata), &metadata)
	if err != nil {
		return nil, err
	}

	callback := j.Get("callback").String()

	// parse url
	callbackURL, err := url.Parse(callback)
	if err != nil {
		return nil, errors.New("callback is not a valid URL")
	}

	// add random nonce to avoid caches
	qs := callbackURL.Query()
	qs.Set("nonce", strconv.FormatInt(time.Now().Unix(), 10))
	callbackURL.RawQuery = qs.Encode()

	return LNURLPayResponse1{
		Tag:            "payRequest",
		Callback:       callback,
		CallbackURL:    callbackURL,
		Metadata:       metadata,
		MaxSendable:    j.Get("maxSendable").Int(),
		MinSendable:    j.Get("minSendable").Int(),
		CommentAllowed: j.Get("commentAllowed").Int(),
	}, nil
}

type Metadata struct {
	Encoded string

	Description     string
	LongDescription string
	Image           struct {
		DataURI string
		Bytes   []byte
		Ext     string
	}
	LightningAddress string
	IsEmail          bool

	PayerIDs struct {
		FreeName         bool
		PubKey           bool
		LightningAddress bool
		Email            bool
		KeyAuth          struct {
			Allowed   bool
			Mandatory bool
			K1        string
		}
	}
}

func (m *Metadata) UnmarshalJSON(src []byte) error {
	m.Encoded = string(src)

	var array []interface{}
	if err := json.Unmarshal(src, &array); err != nil {
		return err
	}

	for _, item := range array {
		entry, _ := item.([]interface{})
		if len(entry) <= 1 {
			continue
		}

		switch entry[0] {
		case "text/plain":
			m.Description, _ = entry[1].(string)
		case "text/long-desc":
			m.LongDescription, _ = entry[1].(string)
		case "image/png;base64", "image/jpeg;base64":
			k, _ := entry[0].(string)
			v, _ := entry[1].(string)

			m.Image.DataURI = "data:" + k + "," + v
			m.Image.Bytes, _ = base64.StdEncoding.DecodeString(v)
			m.Image.Ext = strings.Split(strings.Split(k, "/")[1], ";")[0]
		case "text/email", "text/identifier":
			m.LightningAddress, _ = entry[1].(string)
			if entry[0].(string) == "text/email" {
				m.IsEmail = true
			}
		case "application/payer-ids":
			for _, ialt := range entry[1:] {
				if alt, ok := ialt.([]interface{}); ok {
					if len(alt) > 0 {
						if tag, ok := alt[0].(string); ok {
							switch tag {
							case "text/plain":
								m.PayerIDs.FreeName = true
							case "application/pubkey":
								m.PayerIDs.PubKey = true
							case "application/lnurl-auth":
								if len(alt) == 3 {
									m.PayerIDs.KeyAuth.Allowed = true
									m.PayerIDs.KeyAuth.Mandatory, _ = alt[1].(bool)
									m.PayerIDs.KeyAuth.K1, _ = alt[2].(string)
								}
							case "text/identifier":
								m.PayerIDs.LightningAddress = true
							case "text/email":
								m.PayerIDs.Email = true
							}
						}
					}
				}
			}
		}
	}

	return nil
}

func (m Metadata) MarshalJSON() ([]byte, error) {
	if m.Encoded != "" {
		return json.Marshal(m.Encoded)
	}

	raw := make([]interface{}, 0, 5)
	raw = append(raw, []string{"text/plain", m.Description})

	if m.LongDescription != "" {
		raw = append(raw, []string{"text/long-desc", m.LongDescription})
	}

	if m.Image.Bytes != nil {
		raw = append(raw, []string{"image/" + m.Image.Ext + ";base64",
			base64.StdEncoding.EncodeToString(m.Image.Bytes)})
	} else if m.Image.DataURI != "" {
		raw = append(raw, strings.SplitN(m.Image.DataURI[5:], ",", 2))
	}

	if m.LightningAddress != "" {
		tag := "text/identifier"
		if m.IsEmail {
			tag = "text/email"
		}
		raw = append(raw, []string{tag, m.LightningAddress})
	}

	payerIDs := []interface{}{"application/payer-ids"}
	if m.PayerIDs.FreeName {
		payerIDs = append(payerIDs, []string{"text/plain"})
	}
	if m.PayerIDs.PubKey {
		payerIDs = append(payerIDs, []string{"application/pubkey"})
	}
	if m.PayerIDs.LightningAddress {
		payerIDs = append(payerIDs, []string{"text/identifier"})
	}
	if m.PayerIDs.Email {
		payerIDs = append(payerIDs, []string{"text/email"})
	}
	if m.PayerIDs.KeyAuth.Allowed {
		payerIDs = append(payerIDs, []interface{}{
			"application/lnurl-auth",
			m.PayerIDs.KeyAuth.Mandatory,
			m.PayerIDs.KeyAuth.K1,
		})
	}
	if len(payerIDs) > 1 {
		raw = append(raw, payerIDs)
	}

	j, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}

	return json.Marshal(string(j))
}

func (m Metadata) Hash() [32]byte {
	j, _ := json.Marshal(m)

	var raw string
	json.Unmarshal(j, &raw)

	return sha256.Sum256([]byte(raw))
}
