package moex

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/servererrors"
	"sync"
)

func (m *moex) FindStocks(ctx context.Context, names []string) []model.MoexFindResult {
	var wg sync.WaitGroup
	results := make([]model.MoexFindResult, len(names))
	sem := make(chan struct{}, m.maxConns)

	for i, name := range names {
		wg.Go(func() {
			sem <- struct{}{}
			defer func() { <-sem }()

			results[i] = m.findStock(ctx, name)
		})
	}

	wg.Wait()
	return results
}

func (m *moex) findStock(ctx context.Context, name string) model.MoexFindResult {
	params := url.Values{}
	params.Add("q", name)
	urlStr := fmt.Sprintf("%s/securities.json?%s", m.baseUrl, params.Encode())
	request, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return model.MoexFindResult{Err: err}
	}

	response, err := m.client.Do(request)
	if err != nil {
		return model.MoexFindResult{Err: err}
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return model.MoexFindResult{Err: servererrors.ErrMoexUnhandled}
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return model.MoexFindResult{Err: err}
	}

	var fr findResponse
	err = json.Unmarshal(body, &fr)
	if err != nil {
		return model.MoexFindResult{Err: err}
	}

	res := fr.parseResponse()
	if res.Err != nil {
		return res
	}

	lotSize, price, err := m.getInstrumentInfo(ctx, res.Code)
	if err != nil {
		res.Err = err
		return res
	}
	res.LotSize = lotSize
	res.Price = price

	return res
}

type (
	findResponse struct {
		Securities struct {
			Columns []string `json:"columns"`
			Data    [][]any  `json:"data"`
		} `json:"securities"`
	}
)

func (fr *findResponse) parseResponse() (res model.MoexFindResult) {
	secId := -1
	shortNameId := -1
	typeId := -1

	res = model.MoexFindResult{
		Stock: model.Stock{},
	}

	for i, col := range fr.Securities.Columns {
		if col == "secid" {
			secId = i
		}
		if col == "shortname" {
			shortNameId = i
		}
		if col == "type" {
			typeId = i
		}
	}
	if secId == -1 || shortNameId == -1 || typeId == -1 {
		res.Err = servererrors.ErrMoexUnhandled
		return
	}
	if len(fr.Securities.Data) == 0 {
		res.Err = servererrors.ErrMoexStockNotFound
		return
	}

	for _, row := range fr.Securities.Data {
		if row[typeId] == "common_share" {
			res.Code = row[secId].(string)
			res.Name = row[shortNameId].(string)
			return
		}
	}

	res.Err = servererrors.ErrMoexStockNotFound
	return res
}
