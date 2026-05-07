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

func (m *moex) GetInstrumentsInfo(ctx context.Context, codes []string) []model.MoexInstrumentInfoResult {
	var wg sync.WaitGroup
	results := make([]model.MoexInstrumentInfoResult, len(codes))
	sem := make(chan struct{}, m.maxConns)

	for i, code := range codes {
		wg.Go(func() {
			sem <- struct{}{}
			defer func() { <-sem }()

			lotSize, price, err := m.getInstrumentInfo(ctx, code)
			results[i].LotSize = lotSize
			results[i].Price = price
			results[i].Err = err
		})
	}

	wg.Wait()
	return results
}

func (m *moex) getInstrumentInfo(ctx context.Context, code string) (int, float64, error) {
	const (
		engine = "stock"
		market = "shares"
		board  = "TQBR"
	)
	url := fmt.Sprintf(
		"%s/engines/%s/markets/%s/boards/%s/securities/%s.json",
		m.baseUrl,
		engine,
		market,
		board,
		code,
	)

	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, 0.0, err
	}

	response, err := m.client.Do(request)
	if err != nil {
		return 0, 0.0, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return 0, 0.0, servererrors.ErrMoexUnhandled
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return 0, 0.0, err
	}

	var pr instrumentInfoResponce
	err = json.Unmarshal(body, &pr)
	if err != nil {
		return 0, 0.0, err
	}
	return pr.parseResponse()
}

type (
	instrumentInfoResponce struct {
		Securities struct {
			Columns []string `json:"columns"`
			Data    [][]any  `json:"data"`
		} `json:"securities"`
		MarketData struct {
			Columns []string `json:"columns"`
			Data    [][]any  `json:"data"`
		} `json:"marketdata"`
	}
)

func (ii *instrumentInfoResponce) parseResponse() (int, float64, error) {
	if len(ii.MarketData.Data) == 0 || len(ii.Securities.Data) == 0 {
		return 0, 0.0, servererrors.ErrMoexStockNotFound
	}

	securitiesMap := make(map[string]any, len(ii.Securities.Columns))
	marketDataMap := make(map[string]any, len(ii.MarketData.Columns))

	for i, col := range ii.Securities.Columns {
		securitiesMap[col] = ii.Securities.Data[0][i]
	}
	for i, col := range ii.MarketData.Columns {
		marketDataMap[col] = ii.MarketData.Data[0][i]
	}

	if _, ok := securitiesMap["LOTSIZE"]; !ok {
		return 0, 0.0, servererrors.ErrMoexUnhandled
	}
	lotSize, ok := securitiesMap["LOTSIZE"].(float64)
	if !ok {
		return 0, 0.0, servererrors.ErrMoexUnhandled
	}

	if _, ok := marketDataMap["LAST"]; !ok {
		return 0, 0.0, servererrors.ErrMoexUnhandled
	}
	lastPrice, ok := marketDataMap["LAST"].(float64)
	if ok {
		return int(lotSize), lastPrice, nil
	}

	if _, ok := marketDataMap["MARKETPRICE"]; !ok {
		return 0, 0.0, servererrors.ErrMoexUnhandled
	}
	marketPrice, ok := marketDataMap["MARKETPRICE"].(float64)
	if ok {
		return int(lotSize), marketPrice, nil
	}

	if _, ok := securitiesMap["PREVPRICE"]; !ok {
		return 0, 0.0, servererrors.ErrMoexUnhandled
	}
	prevPrice, ok := securitiesMap["PREVPRICE"].(float64)
	if ok {
		return int(lotSize), prevPrice, nil
	}

	return 0, 0.0, servererrors.ErrMoexUnhandled
}
