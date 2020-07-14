package lnurl

import (
	"encoding/base64"
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
	Callback        string            `json:"callback"`
	CallbackURL     *url.URL          `json:"-"`
	Tag             string            `json:"tag"`
	MaxSendable     int64             `json:"maxSendable"`
	MinSendable     int64             `json:"minSendable"`
	EncodedMetadata string            `json:"metadata"`
	Metadata        [][]string        `json:"-"`
	ParsedMetadata  map[string]string `json:"-"`
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
	var metadata [][]string
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

	// turn metadata into a dictionary
	parsedMetadata := make(map[string]string)
	for _, pair := range metadata {
		parsedMetadata[pair[0]] = pair[1]
	}

	return LNURLPayResponse1{
		Tag:             "payRequest",
		Callback:        callback,
		CallbackURL:     callbackURL,
		EncodedMetadata: strmetadata,
		Metadata:        metadata,
		ParsedMetadata:  parsedMetadata,
		MaxSendable:     j.Get("maxSendable").Int(),
		MinSendable:     j.Get("minSendable").Int(),
	}, nil
}
