package client

import (
	"context"
	"net/http"
	"stocks_calculator/internal/model"
)

func (c *HttpClient) AddStocks(ctx context.Context, stocks []model.Stock) error {
	err := c.doJson(ctx, http.MethodPost, "/addstocks", stocks, nil, true)
	return err
}

func (c *HttpClient) RemoveStocks(ctx context.Context, stocks []model.Stock) error {
	err := c.doJson(ctx, http.MethodDelete, "/removestocks", stocks, nil, true)
	return err
}

func (c *HttpClient) GetStocks(ctx context.Context) ([]model.Stock, error) {
	var stocks []model.Stock
	err := c.doJson(ctx, http.MethodGet, "/getstocks", nil, &stocks, true)
	return stocks, err
}

func (c *HttpClient) FindStocks(ctx context.Context, stockNames []string) ([]model.FoundStock, error) {
	var stocks []model.FoundStock
	err := c.doJson(ctx, http.MethodPost, "/findstocks", stockNames, &stocks, true)
	return stocks, err
}
