package client

import (
	"fmt"
	"slices"
)

type (
	RebalanceBody struct {
		NoSell    bool    `json:"no_sell"`
		ValueDiff float64 `json:"value_diff"`
	}
	RebalanceResponse struct {
		Stocks        []StockInfoShort `json:"stocks"`
		TotalPrevCost float64          `json:"total_prev_cost"`
		CostDiff      float64          `json:"cost_diff"`
	}
)

func GetRebalanceInfo(body RebalanceBody) (RebalanceResponse, error) {
	//temp placeholder for http request

	if body.NoSell && body.ValueDiff <= 0 {
		return RebalanceResponse{}, fmt.Errorf("400 Bad Request")
	}

	stocks := readStock()

	if len(stocks) == 0 {
		return RebalanceResponse{}, fmt.Errorf("Your portfolio is empty")
	}

	var (
		totalPrevCost float64
		costDiff      float64
	)
	for i := range stocks {
		totalPrevCost += float64(stocks[i].CurrentAmount) * stocks[i].CurrentPrice
	}

	response := RebalanceResponse{TotalPrevCost: totalPrevCost}
	if body.ValueDiff > 0 {
		i := 0
		for body.ValueDiff >= 10.0 {
			appended := false
			for j := range response.Stocks {
				if response.Stocks[j].Code == stocks[i].Code {
					response.Stocks[j].Amount += 1
					appended = true
					break
				}
			}
			if !appended {
				response.Stocks = append(response.Stocks, stocks[i].ToShort())
			}

			body.ValueDiff -= 10.0
			costDiff += 10.0
			i++
			if i == len(stocks) {
				i = 0
			}
		}

		response.CostDiff = costDiff
		return response, nil
	}

	if totalPrevCost < (-body.ValueDiff) {
		return RebalanceResponse{}, fmt.Errorf("Total cost is lower that value diff")
	}

	i := 0
	for body.ValueDiff < 0 {
		body.ValueDiff += stocks[i].CurrentPrice * float64(stocks[i].CurrentAmount)
		costDiff -= stocks[i].CurrentPrice * float64(stocks[i].CurrentAmount)

		short := stocks[i].ToShort()
		short.Amount = -stocks[i].CurrentAmount
		response.Stocks = append(response.Stocks, short)

		i++
	}

	response.CostDiff = costDiff
	return response, nil
}

func AcceptRebalance(stocks []StockInfoShort) error {
	//accept portfolio rebalances
	//temp placeholder for http request

	prev := readStock()
	toDelete := []int{}

outer:
	for _, stock := range stocks {
		for j := range prev {
			if stock.Code == prev[j].Code {
				if stock.Amount+prev[j].CurrentAmount < 0 {
					return fmt.Errorf(
						"Cannot sell %d of %q, portfolio only has %d",
						-stock.Amount,
						stock.Name,
						prev[j].CurrentAmount,
					)
				}
				if stock.Amount+prev[j].CurrentAmount == 0 {
					toDelete = append(toDelete, j)
					continue outer
				}
				prev[j].CurrentAmount += stock.Amount
				continue outer
			}
		}
		return fmt.Errorf("Stock not found %q", stock.Name)
	}

	deletedAmount := 0
	for _, ind := range toDelete {
		prev = slices.Delete(prev, ind-deletedAmount, ind+1-deletedAmount)
		deletedAmount++
	}

	writeStock(prev)

	return nil
}
