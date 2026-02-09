package client

import (
	"net/http"
	"time"
)

func newClient() *http.Client {
	return &http.Client{
		Timeout: 10 * time.Second,
	}
}
