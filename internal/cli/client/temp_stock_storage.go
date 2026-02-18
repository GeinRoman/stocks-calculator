package client

import (
	"encoding/json"
	"os"
)

type groupedStock struct {
	Group  string       `json:"group"`
	Stocks []FoundStock `json:"stocks"`
}

const (
	stockfile = "/tmp/stcalc/stock.json"
)

func readStock() (groupedStocks []groupedStock) {
	data, err := os.ReadFile(stockfile)
	if err != nil {
		return []groupedStock{}
	}

	_ = json.Unmarshal(data, &groupedStocks)
	return groupedStocks
}

func writeStock(groupedStocks []groupedStock) {
	data, _ := json.Marshal(groupedStocks)
	os.WriteFile(stockfile, data, 0666)
}
