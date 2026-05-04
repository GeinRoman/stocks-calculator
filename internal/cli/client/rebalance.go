package client

import (
	"context"
	"fmt"
	"net/http"
	"stocks_calculator/internal/model"
)

func (c *HttpClient) GetRebalanceInfo(ctx context.Context, noSell bool, valueDiff float64) (model.RebalanceResponse, error) {
	params := make(map[string]string, 2)
	if noSell {
		params["nosell"] = "t"
	} else {
		params["nosell"] = "f"
	}
	params["valdiff"] = fmt.Sprintf("%.2f", valueDiff)

	var result model.RebalanceResponse
	err := c.doJsonWithQueryParams(
		ctx,
		http.MethodGet,
		"/rebalance",
		nil,
		&result,
		params,
		true,
	)
	return result, err
}
