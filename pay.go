package lnurl

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	decodepay "github.com/fiatjaf/ln-decodepay"
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

type LNURLPayParams struct {
	LNURLResponse
	Callback        string        `json:"callback"`
	Tag             string        `json:"tag"`
	MaxSendable     int64         `json:"maxSendable"`
	MinSendable     int64         `json:"minSendable"`
	EncodedMetadata string        `json:"metadata"`
	CommentAllowed  int64         `json:"commentAllowed"`
	PayerData       PayerDataSpec `json:"payerData,omitempty"`

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
	*PayerDataItemSpec
	K1 string `json:"k1"`
}

type LNURLPayValues struct {
	LNURLResponse
	SuccessAction *SuccessAction `json:"successAction"`
	Routes        []struct{}     `json:"routes"` // always empty
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

func HandlePay(j gjson.Result) (LNURLParams, error) {
	var params LNURLPayParams
	err := json.Unmarshal([]byte(j.Raw), &params)
	if err != nil {
		return nil, err
	}

	// parse metadata
	strmetadata := j.Get("metadata").String()
	var metadata Metadata
	var array []interface{}
	if err := json.Unmarshal([]byte(params.EncodedMetadata), &array); err != nil {
		return params, err
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
	callback := j.Get("callback").String()
	callbackURL, err := url.Parse(callback)
	if err != nil {
		return nil, errors.New("callback is not a valid URL")
	}

	// add random nonce to avoid caches
	qs := callbackURL.Query()
	qs.Set("nonce", strconv.FormatInt(time.Now().Unix(), 10))
	callbackURL.RawQuery = qs.Encode()

	// unmarshal payerdata
	var payerData PayerDataSpec
	json.Unmarshal([]byte(j.Get("payerData").String()), &payerData)

	return LNURLPayParams{
		Tag:             "payRequest",
		Callback:        callbackURL.String(),
		EncodedMetadata: strmetadata,
		Metadata:        metadata,
		MaxSendable:     j.Get("maxSendable").Int(),
		MinSendable:     j.Get("minSendable").Int(),
		CommentAllowed:  j.Get("commentAllowed").Int(),
		PayerData:       payerData,
	}, nil
}

func (params LNURLPayParams) Call(
	msats int64,
	comment string,
	payerdata *PayerDataValues,
) (*LNURLPayValues, error) {
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

	callback := params.CallbackURL()

	qs := callback.Query()
	qs.Set("amount", strconv.FormatInt(msats, 10))

	if comment != "" {
		qs.Set("comment", comment)
	}

	var payerdataJSON string
	if params.PayerData.Exists() && payerdata != nil {
		j, _ := json.Marshal(payerdata)
		payerdataJSON = string(j)
		qs.Set("payerdata", payerdataJSON)
	}

	callback.RawQuery = qs.Encode()

	resp, err := Client.Get(callback.String())
	if err != nil {
		return nil, fmt.Errorf("http error calling '%s': %w",
			callback.String(), err)
	}
	defer resp.Body.Close()

	var values LNURLPayValues
	if err := json.NewDecoder(resp.Body).Decode(&values); err != nil {
		return nil, fmt.Errorf("got invalid JSON from '%s': %w",
			callback.String(), err)
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

	var hhash [32]byte
	if payerdata != nil && params.PayerData.Exists() {
		hhash = params.HashWithPayerData(payerdataJSON)
	} else {
		hhash = params.HashMetadata()
	}

	if inv.DescriptionHash != hex.EncodeToString(hhash[:]) {
		return nil, fmt.Errorf("wrong description_hash (expected %s, got %s)",
			hex.EncodeToString(hhash[:]),
			inv.DescriptionHash,
		)
	}

	if int64(inv.MSatoshi) != msats {
		return nil, fmt.Errorf("got invoice with wrong amount (wanted %d, got %d)",
			msats,
			inv.MSatoshi,
		)
	}

	return &values, nil
}

func (params LNURLPayParams) MetadataEncoded() string {
	if params.EncodedMetadata == "" {
		raw := make([]interface{}, 0, 5)
		raw = append(raw, []string{"text/plain", params.Metadata.Description})

		if params.Metadata.LongDescription != "" {
			raw = append(raw, []string{"text/long-desc", params.Metadata.LongDescription})
		}

		if params.Metadata.Image.Bytes != nil {
			raw = append(raw, []string{"image/" + params.Metadata.Image.Ext + ";base64",
				base64.StdEncoding.EncodeToString(params.Metadata.Image.Bytes)})
		} else if params.Metadata.Image.DataURI != "" {
			raw = append(raw, strings.SplitN(params.Metadata.Image.DataURI[5:], ",", 2))
		}

		if params.Metadata.LightningAddress != "" {
			tag := "text/identifier"
			if params.Metadata.IsEmail {
				tag = "text/email"
			}
			raw = append(raw, []string{tag, params.Metadata.LightningAddress})
		}

		enc, _ := json.Marshal(raw)
		params.EncodedMetadata = string(enc)
	}

	return params.EncodedMetadata
}

func (params LNURLPayParams) HashMetadata() [32]byte {
	return sha256.Sum256([]byte(params.MetadataEncoded()))
}

func (params LNURLPayParams) HashWithPayerData(payerDataJSON string) [32]byte {
	metadataPlusPayerData := params.MetadataEncoded() + payerDataJSON
	return sha256.Sum256([]byte(metadataPlusPayerData))
}
