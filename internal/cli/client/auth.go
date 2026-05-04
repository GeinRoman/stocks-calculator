package client

import (
	"context"
	"net/http"
	"stocks_calculator/internal/model"
)

func (c *HttpClient) Login(ctx context.Context, data model.AuthModel) (model.AuthResponse, error) {
	var result model.AuthResponse
	err := c.doJson(ctx, http.MethodPost, "/login", data, &result, false)

	return result, err
}

func (c *HttpClient) CreateUser(ctx context.Context, data model.AuthModel) (model.AuthResponse, error) {
	var result model.AuthResponse
	err := c.doJson(ctx, http.MethodPost, "/createuser", data, &result, false)

	return result, err
}
