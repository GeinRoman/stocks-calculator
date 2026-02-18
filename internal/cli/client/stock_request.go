package client

import "fmt"

type (
	FindStockBody struct {
		Stocks []string `json:"stocks"`
		Group  string   `json:"group"`
	}
	FoundStock struct {
		Name         string  `json:"name"`
		Code         string  `json:"code"`
		CurrentPrice float64 `json:"current_price"`
		PrevAmount   int     `json:"prev_amount"`
		LotSize      int     `json:"lot_size"`
		ErrorMsg     string  `json:"error_msg"`
	}
	FindStockResponse struct {
		GroupFound  string       `json:"group_found"`
		FoundStocks []FoundStock `json:"found_stocks"`
	}

	AddStockBody struct {
		Stocks []StockBody `json:"stocks"`
		Group  string      `json:"group"`
	}

	StockBody struct {
		Name   string `json:"name"`
		Code   string `json:"code"`
		Amount int    `json:"amount"`
	}
)

func FindStock(body FindStockBody) (FindStockResponse, error) {
	//temp stock finding functionality
	//placeholder for http request
	prevStocks := readStock()
	if !groupExists(&body.Group) {
		return FindStockResponse{}, fmt.Errorf("Group %q was not found", body.Group)
	}

	found := []FoundStock{}
outer:
	for _, s := range body.Stocks {
		for _, g := range prevStocks {
			if body.Group != g.Group {
				continue
			}
			for _, ps := range g.Stocks {
				if s == ps.Name {
					found = append(found, ps)
					continue outer
				}
			}
		}
		found = append(found, FoundStock{s, fmt.Sprintf("CodeOf(%s)", s), 10.0, 0, 1, ""})
	}

	return FindStockResponse{body.Group, found}, nil
}

func AddStock(body AddStockBody) error {
	//temp stock saving functionality
	//placeholder for http request

	prevStocks := readStock()
	if !groupExists(&body.Group) {
		return fmt.Errorf("Group %q was not found", body.Group)
	}

outer:
	for _, s := range body.Stocks {
		for i, g := range prevStocks {
			if g.Group != body.Group {
				continue
			}
			for j, ps := range g.Stocks {
				if s.Code == ps.Code {
					prevStocks[i].Stocks[j].PrevAmount += s.Amount
					continue outer
				}
			}
			prevStocks[i].Stocks = append(prevStocks[i].Stocks, FoundStock{
				Name:         s.Name,
				Code:         s.Code,
				CurrentPrice: 10.0,
				PrevAmount:   s.Amount,
				LotSize:      1,
				ErrorMsg:     "",
			})
			continue outer
		}
		prevStocks = append(prevStocks, groupedStock{Group: body.Group, Stocks: []FoundStock{{
			Name:         s.Name,
			Code:         s.Code,
			CurrentPrice: 10.0,
			PrevAmount:   s.Amount,
			LotSize:      1,
			ErrorMsg:     "",
		}}})
	}

	writeStock(prevStocks)
	return nil
}
