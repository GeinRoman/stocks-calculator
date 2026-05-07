package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"stocks_calculator/internal/cli/config"
	"stocks_calculator/internal/model"
	"time"
)

// if reqBody is nil request body will be empty
func (c *HttpClient) doJson(
	ctx context.Context,
	method string,
	path string,
	reqBody any,
	respBody any,
	needsAuth bool,
) error {
	return c.doJsonWithQueryParams(ctx, method, path, reqBody, respBody, nil, needsAuth)
}

// if reqBody is nil request body will be empty
func (c *HttpClient) doJsonWithQueryParams(
	ctx context.Context,
	method string,
	path string,
	reqBody any,
	respBody any,
	params map[string]string,
	needsAuth bool,
) error {
	var bodyReader io.Reader

	if reqBody != nil {
		b, err := json.Marshal(reqBody)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(b)
	}

	urlStr := fmt.Sprintf("%s%s", c.baseUrl, path)

	if params != nil {
		queryParams := url.Values{}
		for k, v := range params {
			queryParams.Add(k, v)
		}
		urlStr += fmt.Sprintf("?%s", queryParams.Encode())
	}

	req, err := http.NewRequestWithContext(ctx, method, urlStr, bodyReader)
	if err != nil {
		return err
	}

	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if needsAuth {
		if err := c.validateToken(ctx); err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.token.AccessToken)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d: %s", resp.StatusCode, string(b))
	}

	if respBody != nil {
		if err := json.Unmarshal(b, respBody); err != nil {
			return err
		}
	}

	return nil
}

func (c *HttpClient) validateToken(ctx context.Context) error {
	if c.token.AccessToken != "" && time.Now().Before(c.token.ExpiresAt) {
		return nil
	}

	newToken, err := c.refreshToken(ctx)
	if err != nil {
		return err
	}

	c.token = newToken
	return config.UpdateAccessToken(newToken)
}

func (c *HttpClient) refreshToken(ctx context.Context) (model.Token, error) {
	var result model.Token
	err := c.doJson(ctx, http.MethodPost, "/reftoken", c.refToken, &result, false)

	return result, err
}
