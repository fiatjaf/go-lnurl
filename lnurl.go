package lnurl

import (
	"net/http"
	"strings"
	"time"
)

var TorClient *http.Client
var Client = &http.Client{
	Timeout: 5 * time.Second,
}

var actualClient = &http.Client{
	Transport: onioncapabletransport{},
}

type onioncapabletransport struct{}

func (_ onioncapabletransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if strings.HasSuffix(r.URL.Host, ".onion") && TorClient != nil {
		return TorClient.Transport.RoundTrip(r)
	}

	return Client.Transport.RoundTrip(r)
}
