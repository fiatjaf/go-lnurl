package lnurl

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"

	decodepay "github.com/nbd-wtf/ln-decodepay"
)

var (
	f     bool  = false
	t     bool  = true
	FALSE *bool = &f
	TRUE  *bool = &t
)

func CallPay(
	metadata string,
	callback *url.URL,
	msats int64,
	comment string,
	payerdata *PayerDataValues,
) (*LNURLPayValues, error) {
	qs := callback.Query()
	qs.Set("amount", strconv.FormatInt(msats, 10))

	if comment != "" {
		qs.Set("comment", comment)
	}

	var payerdataJSON string
	if payerdata != nil {
		j, _ := json.Marshal(payerdata)
		payerdataJSON = string(j)
		qs.Set("payerdata", payerdataJSON)
	}

	callback.RawQuery = qs.Encode()
	resp, err := actualClient.Get(callback.String())
	if err != nil {
		return nil, fmt.Errorf("http error calling '%s': %w", callback.String(), err)
	}
	defer resp.Body.Close()

	var values LNURLPayValues
	b, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(b, &values); err != nil {
		return nil, fmt.Errorf("got invalid JSON from '%s': %w (%s)",
			callback.String(), err, string(b))
	}

	if values.Status == "ERROR" {
		return nil, LNURLErrorResponse{
			Status: values.Status,
			Reason: values.Reason,
			URL:    callback,
		}
	}

	inv, err := decodepay.Decodepay(values.PR)
	if err != nil {
		return nil, fmt.Errorf("error parsing invoice '%s': %w", values.PR, err)
	}

	values.ParsedInvoice = inv
	values.PayerDataJSON = payerdataJSON

	if int64(inv.MSatoshi) != msats {
		return nil, fmt.Errorf("got invoice with wrong amount (wanted %d, got %d)",
			msats,
			inv.MSatoshi,
		)
	}

	return &values, nil
}

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

type LNURLPayParams struct {
	LNURLResponse
	Callback        string         `json:"callback"`
	Tag             string         `json:"tag"`
	MaxSendable     int64          `json:"maxSendable"`
	MinSendable     int64          `json:"minSendable"`
	EncodedMetadata string         `json:"metadata"`
	CommentAllowed  int64          `json:"commentAllowed"`
	PayerData       *PayerDataSpec `json:"payerData,omitempty"`

	Metadata Metadata `json:"-"`
}

type Metadata struct {
	Description     string
	LongDescription string
	Image           struct {
		DataURI string
		Bytes   []byte
		Ext     string
	}
	LightningAddress string
	IsEmail          bool
}

type PayerDataSpec struct {
	FreeName         *PayerDataItemSpec    `json:"name"`
	PubKey           *PayerDataItemSpec    `json:"pubkey"`
	LightningAddress *PayerDataItemSpec    `json:"identifier"`
	Email            *PayerDataItemSpec    `json:"email"`
	KeyAuth          *PayerDataKeyAuthSpec `json:"auth"`
}

type PayerDataItemSpec struct {
	Mandatory bool `json:"mandatory"`
}

type PayerDataKeyAuthSpec struct {
	Mandatory bool   `json:"mandatory"`
	K1        string `json:"k1"`
}

type LNURLPayValues struct {
	LNURLResponse
	SuccessAction *SuccessAction `json:"successAction"`
	Routes        interface{}    `json:"routes"` // ignored
	PR            string         `json:"pr"`
	Disposable    *bool          `json:"disposable,omitempty"`

	ParsedInvoice decodepay.Bolt11 `json:"-"`
	PayerDataJSON string           `json:"-"`
}

type PayerDataValues struct {
	FreeName         string                  `json:"name,omitempty"`
	PubKey           string                  `json:"pubkey,omitempty"`
	LightningAddress string                  `json:"identifier,omitempty"`
	Email            string                  `json:"email,omitempty"`
	KeyAuth          *PayerDataKeyAuthValues `json:"auth,omitempty"`
}

type PayerDataKeyAuthValues struct {
	K1  string `json:"k1"`
	Sig string `json:"sig"`
	Key string `json:"key"`
}

type SuccessAction struct {
	Tag         string `json:"tag"`
	Description string `json:"description,omitempty"`
	URL         string `json:"url,omitempty"`
	Message     string `json:"message,omitempty"`
	Ciphertext  string `json:"ciphertext,omitempty"`
	IV          string `json:"iv,omitempty"`
}

func (params LNURLPayParams) CallbackURL() *url.URL {
	parsed, _ := url.Parse(params.Callback)
	return parsed
}

func (s PayerDataSpec) Exists() bool {
	return s.FreeName != nil || s.PubKey != nil || s.LightningAddress != nil || s.Email != nil || s.KeyAuth != nil
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

func (_ LNURLPayParams) LNURLKind() string { return "lnurl-pay" }

func HandlePay(raw []byte) (LNURLParams, error) {
	var params LNURLPayParams
	err := json.Unmarshal(raw, &params)
	if err != nil {
		return nil, err
	}

	if err := params.Normalize(); err != nil {
		return params, err
	}

	return params, nil
}

func (params *LNURLPayParams) Normalize() error {
	// parse metadata
	var array []interface{}
	if err := json.Unmarshal([]byte(params.EncodedMetadata), &array); err != nil {
		return err
	}
	for _, item := range array {
		entry, _ := item.([]interface{})
		if len(entry) <= 1 {
			continue
		}

		switch entry[0] {
		case "text/plain":
			params.Metadata.Description, _ = entry[1].(string)
		case "text/long-desc":
			params.Metadata.LongDescription, _ = entry[1].(string)
		case "image/png;base64", "image/jpeg;base64":
			k, _ := entry[0].(string)
			v, _ := entry[1].(string)

			params.Metadata.Image.DataURI = "data:" + k + "," + v
			params.Metadata.Image.Bytes, _ = base64.StdEncoding.DecodeString(v)
			params.Metadata.Image.Ext = strings.Split(strings.Split(k, "/")[1], ";")[0]
		case "text/email", "text/identifier":
			params.Metadata.LightningAddress, _ = entry[1].(string)
			if entry[0].(string) == "text/email" {
				params.Metadata.IsEmail = true
			}
		}
	}

	// parse url
	callbackURL, err := url.Parse(params.Callback)
	if err != nil {
		return errors.New("callback is not a valid URL")
	}

	// add random nonce to avoid caches
	qs := callbackURL.Query()
	qs.Set("__n", strconv.FormatInt(time.Now().Unix(), 10))
	callbackURL.RawQuery = qs.Encode()
	params.Callback = callbackURL.String()

	return nil
}

func (params LNURLPayParams) Call(
	msats int64,
	comment string,
	payerdata *PayerDataValues,
) (*LNURLPayValues, error) {
	if params.PayerData == nil || !params.PayerData.Exists() {
		payerdata = nil
	} else {
		if params.PayerData.Email != nil &&
			params.PayerData.Email.Mandatory &&
			(payerdata == nil || payerdata.Email == "") {
			return nil, fmt.Errorf("email is mandatory")
		}
		if params.PayerData.LightningAddress != nil &&
			params.PayerData.LightningAddress.Mandatory &&
			(payerdata == nil || payerdata.LightningAddress == "") {
			return nil, fmt.Errorf("lightning address is mandatory")
		}
		if params.PayerData.FreeName != nil &&
			params.PayerData.FreeName.Mandatory &&
			(payerdata == nil || payerdata.FreeName == "") {
			return nil, fmt.Errorf("name is mandatory")
		}
		if params.PayerData.PubKey != nil &&
			params.PayerData.PubKey.Mandatory &&
			(payerdata == nil || payerdata.PubKey == "") {
			return nil, fmt.Errorf("pubkey is mandatory")
		}
		if params.PayerData.KeyAuth != nil &&
			params.PayerData.KeyAuth.Mandatory &&
			(payerdata == nil || payerdata.KeyAuth == nil) {
			return nil, fmt.Errorf("auth is mandatory")
		}
	}

	return CallPay(
		params.MetadataEncoded(),
		params.CallbackURL(),
		msats,
		comment,
		payerdata,
	)
}

func (params LNURLPayParams) MetadataEncoded() string {
	if params.EncodedMetadata == "" {
		params.EncodedMetadata = params.Metadata.Encode()
	}

	return params.EncodedMetadata
}

func (metadata Metadata) Encode() string {
	raw := make([]interface{}, 0, 5)
	raw = append(raw, []string{"text/plain", metadata.Description})

	if metadata.LongDescription != "" {
		raw = append(raw, []string{"text/long-desc", metadata.LongDescription})
	}

	if metadata.Image.Bytes != nil {
		raw = append(raw, []string{
			"image/" + metadata.Image.Ext + ";base64",
			base64.StdEncoding.EncodeToString(metadata.Image.Bytes),
		})
	} else if metadata.Image.DataURI != "" {
		raw = append(raw, strings.SplitN(metadata.Image.DataURI[5:], ",", 2))
	}

	if metadata.LightningAddress != "" {
		tag := "text/identifier"
		if metadata.IsEmail {
			tag = "text/email"
		}
		raw = append(raw, []string{tag, metadata.LightningAddress})
	}

	enc, _ := json.Marshal(raw)
	return string(enc)
}
