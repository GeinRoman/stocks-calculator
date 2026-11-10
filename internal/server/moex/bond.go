package moex

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"stocks_calculator/internal/model"
	"strconv"
	"strings"
	"sync"
	"time"
)

var errUnsupportedBondCurrency = errors.New("unsupported bond currency")

type bondTable struct {
	Columns []string `json:"columns"`
	Data    [][]any  `json:"data"`
}

func (t bondTable) rows(required ...string) ([]map[string]any, error) {
	indices := make(map[string]int, len(t.Columns))
	for i, name := range t.Columns {
		name = strings.ToLower(name)
		if _, exists := indices[name]; exists {
			return nil, fmt.Errorf("duplicate MOEX column %s", name)
		}
		indices[name] = i
	}
	for _, name := range required {
		if _, ok := indices[strings.ToLower(name)]; !ok {
			return nil, fmt.Errorf("missing MOEX column %s", name)
		}
	}
	rows := make([]map[string]any, 0, len(t.Data))
	for _, values := range t.Data {
		if len(values) != len(t.Columns) {
			return nil, fmt.Errorf("invalid MOEX row length")
		}
		row := make(map[string]any, len(indices))
		for name, i := range indices {
			row[name] = values[i]
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func (m *moex) bondRequest(ctx context.Context, path string, query url.Values, out any) error {
	u := m.baseUrl + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	resp, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("MOEX HTTP status %d", resp.StatusCode)
	}
	return json.NewDecoder(io.LimitReader(resp.Body, 8<<20)).Decode(out)
}

func (m *moex) GetBondsInfo(ctx context.Context, codes []string) []model.MoexBondInfoResult {
	results := make([]model.MoexBondInfoResult, len(codes))
	var wg sync.WaitGroup
	limit := m.maxConns
	if limit < 1 {
		limit = 1
	}
	sem := make(chan struct{}, limit)
	for i, code := range codes {
		wg.Go(func() {
			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				results[i].Err = ctx.Err()
				return
			}
			defer func() { <-sem }()
			results[i].Bond, results[i].Err = m.getBondInfo(ctx, code)
		})
	}
	wg.Wait()
	return results
}

func rub(v any) bool {
	s, ok := v.(string)
	return ok && (s == "RUB" || s == "SUR")
}

func (m *moex) getBondInfo(ctx context.Context, code string) (model.Bond, error) {
	bond := model.Bond{Code: code}
	if code == "" {
		return bond, fmt.Errorf("empty bond code")
	}
	var discovery struct {
		Boards bondTable `json:"boards"`
	}
	if err := m.bondRequest(ctx, "/securities/"+url.PathEscape(code)+".json", nil, &discovery); err != nil {
		return bond, err
	}
	boards, err := discovery.Boards.rows("boardid", "engine", "market", "is_traded", "is_primary", "unit", "currencyid")
	if err != nil {
		return bond, err
	}
	board := ""
	for _, row := range boards {
		engine, _ := row["engine"].(string)
		market, _ := row["market"].(string)
		traded, _ := row["is_traded"].(float64)
		primary, _ := row["is_primary"].(float64)
		if engine != "stock" || market != "bonds" || traded != 1 || primary != 1 {
			continue
		}
		unit, _ := row["unit"].(string)
		if unit != "%" {
			return bond, fmt.Errorf("unsupported bond board unit")
		}
		if currency, ok := row["currencyid"].(string); !ok || currency == "" {
			return bond, fmt.Errorf("invalid bond board currency")
		} else if !rub(currency) {
			return bond, errUnsupportedBondCurrency
		}
		id, ok := row["boardid"].(string)
		if !ok || id == "" || board != "" {
			return bond, fmt.Errorf("invalid primary bond board")
		}
		board = id
	}
	if board == "" {
		return bond, fmt.Errorf("primary traded bond board not found")
	}
	var snapshot struct {
		Securities bondTable `json:"securities"`
		MarketData bondTable `json:"marketdata"`
	}
	path := "/engines/stock/markets/bonds/boards/" + url.PathEscape(board) + "/securities/" + url.PathEscape(code) + ".json"
	if err := m.bondRequest(ctx, path, nil, &snapshot); err != nil {
		return bond, err
	}
	securities, err := snapshot.Securities.rows("LOTSIZE", "FACEVALUEONSETTLEDATE", "ACCRUEDINT", "MATDATE", "FACEUNIT")
	if err != nil || len(securities) != 1 {
		return bond, fmt.Errorf("invalid bond securities snapshot")
	}
	market, err := snapshot.MarketData.rows()
	if err != nil || len(market) != 1 {
		return bond, fmt.Errorf("invalid bond market snapshot")
	}
	s := securities[0]
	if isin := s["isin"]; isin != nil {
		var ok bool
		bond.ISIN, ok = isin.(string)
		if !ok {
			return bond, fmt.Errorf("invalid bond ISIN")
		}
	}
	if currency, ok := s["faceunit"].(string); !ok || currency == "" {
		return bond, fmt.Errorf("invalid bond face currency")
	} else if !rub(currency) {
		return bond, errUnsupportedBondCurrency
	}
	if currency, exists := s["currencyid"]; exists {
		if currency, ok := currency.(string); !ok || currency == "" {
			return bond, fmt.Errorf("invalid bond trading currency")
		} else if !rub(currency) {
			return bond, errUnsupportedBondCurrency
		}
	}
	face, ok := s["facevalueonsettledate"].(float64)
	if !ok || face <= 0 {
		return bond, fmt.Errorf("invalid settlement face value")
	}
	accrued, ok := s["accruedint"].(float64)
	if !ok || accrued < 0 {
		return bond, fmt.Errorf("invalid accrued interest")
	}
	lot, ok := s["lotsize"].(float64)
	if !ok || lot < 1 || lot != math.Trunc(lot) || lot >= float64(math.MaxInt) {
		return bond, fmt.Errorf("invalid bond lot size")
	}
	maturity, ok := s["matdate"].(string)
	if !ok || !validBondDate(maturity) {
		return bond, fmt.Errorf("invalid bond maturity date")
	}
	var price float64
	for _, value := range []any{market[0]["last"], market[0]["marketprice"], s["prevprice"]} {
		if n, ok := value.(float64); ok && n > 0 {
			price = n
			break
		}
	}
	if price == 0 || math.IsInf(price*face, 0) {
		return bond, fmt.Errorf("invalid bond price")
	}
	bond.CleanPrice = price * face / 100
	bond.AccruedInterest, bond.LotSize, bond.MaturityDate = accrued, int(lot), maturity
	if name, ok := s["shortname"].(string); ok {
		bond.Name = name
	}
	bond.MaturityPayment, err = m.bondMaturityPayment(ctx, code, maturity)
	return bond, err
}

func validBondDate(date string) bool {
	t, err := time.Parse("2006-01-02", date)
	return err == nil && t.Format("2006-01-02") == date
}

func (m *moex) bondMaturityPayment(ctx context.Context, code, maturity string) (*float64, error) {
	start, total := 0, -1
	var payment float64
	var paymentDate string
	for {
		var response struct {
			Amortizations bondTable `json:"amortizations"`
			Cursor        bondTable `json:"amortizations.cursor"`
		}
		path := "/statistics/engines/stock/markets/bonds/bondization/" + url.PathEscape(code) + ".json"
		if err := m.bondRequest(ctx, path, url.Values{"start": {strconv.Itoa(start)}}, &response); err != nil {
			return nil, err
		}
		rows, err := response.Amortizations.rows("amortdate", "value", "faceunit")
		if err != nil {
			return nil, err
		}
		cursor, err := response.Cursor.rows("INDEX", "TOTAL", "PAGESIZE")
		if err != nil || len(cursor) != 1 {
			return nil, fmt.Errorf("invalid amortization cursor")
		}
		index, iok := cursor[0]["index"].(float64)
		count, tok := cursor[0]["total"].(float64)
		pageSize, pok := cursor[0]["pagesize"].(float64)
		if !iok || !tok || !pok || index != float64(start) || count < 1 || count != math.Trunc(count) || count >= float64(math.MaxInt) || pageSize < 1 || pageSize != math.Trunc(pageSize) || float64(len(rows)) > pageSize || len(rows) == 0 || float64(start)+float64(len(rows)) > count {
			return nil, fmt.Errorf("incomplete amortization schedule")
		}
		if total < 0 {
			total = int(count)
		} else if total != int(count) {
			return nil, fmt.Errorf("amortization schedule changed")
		}
		for _, row := range rows {
			date, ok := row["amortdate"].(string)
			if !ok || !validBondDate(date) || !rub(row["faceunit"]) {
				return nil, fmt.Errorf("invalid amortization date or currency")
			}
			value, ok := row["value"].(float64)
			if !ok || value <= 0 {
				return nil, fmt.Errorf("invalid amortization value")
			}
			payment, paymentDate = value, date
		}
		start += len(rows)
		if start == total {
			break
		}
	}
	if total > 1 {
		return nil, nil
	}
	if paymentDate != maturity {
		return nil, fmt.Errorf("principal payment does not match maturity")
	}
	return &payment, nil
}
