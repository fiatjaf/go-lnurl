package lnurl

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/url"
	"strconv"
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
	Callback        string   `json:"callback"`
	CallbackURL     *url.URL `json:"-"`
	Tag             string   `json:"tag"`
	MaxSendable     int64    `json:"maxSendable"`
	MinSendable     int64    `json:"minSendable"`
	EncodedMetadata string   `json:"metadata"`
	Metadata        Metadata `json:"-"`
	CommentAllowed  int64    `json:"commentAllowed"`
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
		Tag:             "payRequest",
		Callback:        callback,
		CallbackURL:     callbackURL,
		EncodedMetadata: strmetadata,
		Metadata:        metadata,
		MaxSendable:     j.Get("maxSendable").Int(),
		MinSendable:     j.Get("minSendable").Int(),
		CommentAllowed:  j.Get("commentAllowed").Int(),
	}, nil
}

type Metadata [][]interface{}

// Description returns the content of text/plain metadata entry.
func (m Metadata) Description() string {
	return m.Entry("text/plain")
}

// LongDescription returns the content of text/long-desc metadata entry.
func (m Metadata) LongDescription() string {
	return m.Entry("text/long-desc")
}

// ImageDataURI returns image in the form data:image/type;base64,... if an image exists
// or an empty string if not.
func (m Metadata) ImageDataURI() string {
	if v := m.Entry("image/png;base64"); v != "" {
		return "data:image/png;base64," + v
	}
	if v := m.Entry("image/jpeg;base64"); v != "" {
		return "data:image/jpeg;base64," + v
	}
	return ""
}

// ImageBytes returns image as bytes, decoded from base64 if an image exists
// or nil if not.
func (m Metadata) ImageBytes() []byte {
	if v := m.Entry("image/png;base64"); v != "" {
		if decoded, err := base64.StdEncoding.DecodeString(v); err == nil {
			return decoded
		}
	}
	if v := m.Entry("image/jpeg;base64"); v != "" {
		if decoded, err := base64.StdEncoding.DecodeString(v); err == nil {
			return decoded
		}
	}
	return nil
}

// ImageExtension returns the file extension for the image, either "png" or "jpeg"
func (m Metadata) ImageExtension() string {
	if v := m.Entry("image/png;base64"); v != "" {
		return "png"
	}
	if v := m.Entry("image/jpeg;base64"); v != "" {
		return "jpeg"
	}
	return ""
}

// LightningAddress returns either text/identifier or text/email
func (m Metadata) LightningAddress() string {
	if identifier := m.Entry("text/identifier"); identifier != "" {
		return identifier
	}
	if email := m.Entry("text/email"); email != "" {
		return email
	}
	return ""
}

// PayerIDs
type PayerIDs struct {
	FreeName         bool
	PubKey           bool
	LightningAddress bool
	Email            bool
	KeyAuth          struct {
		Allowed   bool
		Mandatory bool
		K1        []byte
	}
}

func (m Metadata) AllowedPayerIDs() PayerIDs {
	payerIDs := PayerIDs{}

	for _, entry := range m {
		if len(entry) > 1 && entry[0] == "application/payer-ids" {
			for _, ialt := range entry[1:] {
				if alt, ok := ialt.([]interface{}); ok {
					if len(alt) > 0 {
						if tag, ok := alt[0].(string); ok {
							switch tag {
							case "text/plain":
								payerIDs.FreeName = true
							case "application/pubkey":
								payerIDs.PubKey = true
							case "application/lnurl-auth":
								if len(alt) == 3 {
									payerIDs.KeyAuth.Allowed = true
									payerIDs.KeyAuth.Mandatory, _ = alt[1].(bool)
									k1, _ := alt[2].(string)
									payerIDs.KeyAuth.K1, _ = hex.DecodeString(k1)
								}
							case "text/identifier":
								payerIDs.LightningAddress = true
							case "text/email":
								payerIDs.Email = true
							}
						}
					}
				}
			}
		}
	}

	return payerIDs
}

// Entry returns an arbitrary entry from the metadata array.
// eg.: "video/mp4" or "application/vnd.some-specific-thing-from-a-specific-app".
func (m Metadata) Entry(key string) string {
	for _, entry := range m {
		if len(entry) == 2 && entry[0] == key {
			if v, ok := entry[1].(string); ok {
				return v
			}
		}
	}
	return ""
}
