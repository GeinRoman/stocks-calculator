package app

import (
	"fmt"
	"stocks_calculator/internal/cli/client"
	"strings"
)

type RebalanceOptions struct {
	NoSell   bool
	Deposit  int
	Withdraw int
}

func Rebalance(options RebalanceOptions) (string, error) {
	body := client.RebalanceBody{NoSell: options.NoSell}
	if options.Deposit != 0 {
		body.ValueDiff = float64(options.Deposit)
	}
	if options.Withdraw != 0 {
		body.ValueDiff = float64(-options.Withdraw)
	}

	response, err := client.GetRebalanceInfo(body)
	if err != nil {
		return "", fmt.Errorf("Failed to retrieve rebalance info. (%w)", err)
	}

	fmt.Print(formatRebalanceMessage(response))
	//allready balanced portfolio no additional confiramtions needed
	if len(response.Stocks) == 0 {
		return "", nil
	}

	confirmed, err := confirmation("Do you want to automatically update portfolio info according to provided rebalance?")
	if err != nil {
		return "", fmt.Errorf("Fail to read user input")
	}
	if !confirmed {
		return "Portfolio wasn't automatically updated", nil
	}

	err = client.AcceptRebalance(response.Stocks)
	if err != nil {
		return "", fmt.Errorf("Failed to automatically update portfolio. (%w)", err)
	}

	return "Portfolio was successfully updated", nil
}

func formatRebalanceMessage(resp client.RebalanceResponse) string {
	var msg strings.Builder

	fmt.Fprintf(&msg, "Current Portfolio Value: %.2f RUB\n", resp.TotalPrevCost)

	if resp.CostDiff > 0 {
		fmt.Fprintf(&msg, "Required Deposit: %.2f RUB\n", resp.CostDiff)
	} else if resp.CostDiff < 0 {
		fmt.Fprintf(&msg, "Available Withdrawal: %.2f RUB\n", -resp.CostDiff)
	} else {
		msg.WriteString("No additional capital needed\n")
	}

	msg.WriteString("\n")

	if len(resp.Stocks) == 0 {
		msg.WriteString("Portfolio is already balanced - no trades needed\n")
		return msg.String()
	}

	buys := []client.StockInfoShort{}
	sells := []client.StockInfoShort{}

	for _, stock := range resp.Stocks {
		if stock.Amount > 0 {
			buys = append(buys, stock)
		} else if stock.Amount < 0 {
			sells = append(sells, stock)
		}
	}

	if len(sells) > 0 {
		msg.WriteString("Sell:\n")
		for _, stock := range sells {
			fmt.Fprintf(
				&msg,
				"  - %s (%s): %d lot(s)\n",
				stock.Name,
				stock.Code,
				-stock.Amount,
			)
		}
		msg.WriteString("\n")
	}

	if len(buys) > 0 {
		msg.WriteString("Buy:\n")
		for _, stock := range buys {
			fmt.Fprintf(
				&msg,
				"  + %s (%s): %d lot(s)\n",
				stock.Name,
				stock.Code,
				stock.Amount,
			)
		}
		msg.WriteString("\n")
	}

	return msg.String()
}
