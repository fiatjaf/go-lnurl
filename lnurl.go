package lnurl

import (
	"net/http"
	"time"
)

var Client = &http.Client{Timeout: 5 * time.Second}
