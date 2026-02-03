package client

import (
	"net/http"
	"time"
)

func new() *http.Client {
	return &http.Client{
		Timeout: 10 * time.Second,
	}
}
