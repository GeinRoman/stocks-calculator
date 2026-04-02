package client

import (
	"fmt"
	"slices"
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
	StockInfoShort struct {
		Name    string `json:"name"`
		Code    string `json:"code"`
		GroupId uint   `json:"group_id"`
		Amount  int    `json:"amount"`
	}
)

func (info *StockInfo) ToShort() StockInfoShort {
	return StockInfoShort{
		Name:    info.Name,
		Code:    info.Code,
		GroupId: info.GroupId,
		Amount:  1,
	}
}

func AddStock(stocks []StockInfoShort) error {
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

func RemoveStock(stocks []StockInfoShort) error {
	//temp stock removing functionality
	//placeholder for http request

	prevStocks := readStock()

	toDelete := []int{}
outer:
	for i := range stocks {
		for j := range prevStocks {
			if stocks[i].Code == prevStocks[j].Code {
				if prevStocks[j].CurrentAmount > stocks[i].Amount {
					prevStocks[j].CurrentAmount -= stocks[i].Amount
				} else {
					toDelete = append(toDelete, j)
				}
				continue outer
			}
		}
		return fmt.Errorf("Failed to fild stock %q, code: %q to delete", stocks[i].Name, stocks[i].Code)
	}

	deletedAmount := 0
	for _, ind := range toDelete {
		prevStocks = slices.Delete(prevStocks, ind-deletedAmount, ind+1-deletedAmount)
		deletedAmount++
	}

	writeStock(prevStocks)
	return nil
}

type (
	StocksInfo struct {
		Stocks []StockInfo `json:"stocks"`
		Groups []Group     `json:"groups"`
	}
)

func GetStocksInfo() (StocksInfo, error) {
	//temp stock info functionality
	//placeholder for http request

	return StocksInfo{readStock(), readGroups()}, nil
}
