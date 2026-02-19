package client

import (
	"fmt"
)

type (
	StockInfo struct {
		Name          string  `json:"name"`
		Code          string  `json:"code"`
		GroupId       uint    `json:"group_id,omitempty"`
		CurrentPrice  float64 `json:"current_price"`
		CurrentAmount int     `json:"current_amount"`
		LotSize       int     `json:"lot_size"`
		ErrorMsg      string  `json:"error_msg"`
	}
)

func FindStock(stocks []string) ([]StockInfo, error) {
	//temp stock finding functionality
	//placeholder for http request
	prevStocks := readStock()

	found := []StockInfo{}

outer:
	for i := range stocks {
		for j := range prevStocks {
			if stocks[i] == prevStocks[j].Name || stocks[i] == prevStocks[j].Code {
				found = append(found, prevStocks[j])
				continue outer
			}
		}
		found = append(found, StockInfo{
			Name:          stocks[i],
			Code:          fmt.Sprintf("CodeOf(%s)", stocks[i]),
			GroupId:       0,
			CurrentPrice:  10.0,
			CurrentAmount: 0,
			LotSize:       1,
			ErrorMsg:      "",
		})
	}

	return found, nil
}

type (
	StockShortInfo struct {
		Name    string `json:"name"`
		Code    string `json:"code"`
		GroupId uint   `json:"group_id"`
		Amount  int    `json:"amount"`
	}
)

func AddStock(stocks []StockShortInfo) error {
	//temp stock saving functionality
	//placeholder for http request

	prevStocks := readStock()

outer:
	for i := range stocks {
		for j := range prevStocks {
			if stocks[i].Code == prevStocks[j].Code {
				prevStocks[j].CurrentAmount += stocks[i].Amount
				continue outer
			}
		}

		prevStocks = append(prevStocks, StockInfo{
			Name:          stocks[i].Name,
			Code:          stocks[i].Code,
			GroupId:       stocks[i].GroupId,
			CurrentPrice:  10.0,
			CurrentAmount: stocks[i].Amount,
			LotSize:       1,
			ErrorMsg:      "",
		})
	}

	writeStock(prevStocks)
	return nil
}
