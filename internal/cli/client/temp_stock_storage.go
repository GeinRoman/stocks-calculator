package client

import (
	"encoding/json"
	"os"
)

const (
	stockfile = "/tmp/stcalc/stock.json"
)

func readStock() (stocks []StockInfo) {
	data, err := os.ReadFile(stockfile)
	if err != nil {
		return []StockInfo{}
	}

	_ = json.Unmarshal(data, &stocks)
	return stocks
}

func writeStock(stocks []StockInfo) {
	data, _ := json.Marshal(stocks)
	os.WriteFile(stockfile, data, 0666)
}
