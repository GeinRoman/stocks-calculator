package client

import (
	"net/http"
	"stocks_calculator/internal/model"
	"time"
)

type HttpClient struct {
	client   *http.Client
	baseUrl  string
	token    model.Token
	refToken string
}

func New(baseURL string, token model.Token, refToken string) *HttpClient {
	return &HttpClient{
		baseUrl:  baseURL,
		token:    token,
		refToken: refToken,
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}
