package moex

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/servererrors"
	"sync"
)

func (m *moex) GetPrices(ctx context.Context, codes []string) []model.MoexPriceResult {
	var wg sync.WaitGroup
	results := make([]model.MoexPriceResult, len(codes))
	sem := make(chan struct{}, m.maxConns)

	for i, code := range codes {
		wg.Go(func() {
			sem <- struct{}{}
			defer func() { <-sem }()

			price, err := m.getPrice(ctx, code)
			results[i].Price = price
			results[i].Err = err
		})
	}

	wg.Wait()
	return results
}

func (m *moex) getPrice(ctx context.Context, code string) (float64, error) {
	url := fmt.Sprintf("%s/securities/%s/aggregates.json", m.baseUrl, code)
	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, err
	}

	response, err := m.client.Do(request)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return 0, servererrors.ErrMoexUnhandled
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return 0, err
	}

	var pr priceResponse
	err = json.Unmarshal(body, &pr)
	if err != nil {
		return 0, err
	}
	return pr.parseResponse()
}

type (
	priceResponse struct {
		Aggregates struct {
			Columns []string `json:"columns"`
			Data    [][]any  `json:"data"`
		} `json:"aggregates"`
	}
)

func (fr *priceResponse) parseResponse() (float64, error) {
	marketNameInd := -1
	valueInd := -1
	volumeInd := -1

	for i, col := range fr.Aggregates.Columns {
		if col == "market_name" {
			marketNameInd = i
		}
		if col == "value" {
			valueInd = i
		}
		if col == "volume" {
			volumeInd = i
		}
	}
	if marketNameInd == -1 || valueInd == -1 || volumeInd == -1 {
		return 0, servererrors.ErrMoexUnhandled
	}
	if len(fr.Aggregates.Data) == 0 {
		return 0, servererrors.ErrMoexStockNotFound
	}

	for _, row := range fr.Aggregates.Data {
		if row[marketNameInd] == "shares" {
			price := row[valueInd].(float64) / row[volumeInd].(float64)
			return price, nil
		}
	}

	return 0, servererrors.ErrMoexStockNotFound
}
