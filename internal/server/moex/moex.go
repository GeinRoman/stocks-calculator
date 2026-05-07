package moex

import (
	"net"
	"net/http"
	"time"
)

type moex struct {
	client   *http.Client
	maxConns int
	baseUrl  string
}

func New(maxConns int) *moex {

	dialer := &net.Dialer{
		Timeout:   3 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	transport := &http.Transport{
		DialContext:           dialer.DialContext,
		MaxIdleConns:          maxConns,
		MaxIdleConnsPerHost:   maxConns,
		MaxConnsPerHost:       maxConns,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ForceAttemptHTTP2:     true,
	}

	return &moex{
		baseUrl:  "https://iss.moex.com/iss",
		maxConns: maxConns,
		client: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
	}
}
