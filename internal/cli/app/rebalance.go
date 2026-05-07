package app

import (
	"context"
	"fmt"
	"stocks_calculator/internal/cli/client"
	"stocks_calculator/internal/cli/config"
	"stocks_calculator/internal/model"
	"strings"
	"time"
)

type RebalanceOptions struct {
	NoSell   bool
	Deposit  int
	Withdraw int
}

func Rebalance(options RebalanceOptions) (string, error) {
	httpClient := client.New(config.Url(), config.UserConfig.Token, config.UserConfig.RefToken)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var valueDiff float64
	if options.Deposit != 0 {
		valueDiff = float64(options.Deposit)
	}
	if options.Withdraw != 0 {
		valueDiff = float64(-options.Withdraw)
	}
	response, err := httpClient.GetRebalanceInfo(ctx, options.NoSell, valueDiff)
	if err != nil {
		return "", fmt.Errorf("Failed to retrieve rebalance info. (%w)", err)
	}

	originalStocks, err := httpClient.GetStocks(ctx)
	if err != nil {
		return "", err
	}

	fmt.Print(fprintGeneralPortfolioInfo(response))

	buys, sells := sortRebalance(response, originalStocks)
	if len(buys) == 0 && len(sells) == 0 {
		return "Portfolio is already balanced - no trades needed\n", nil
	}
	fmt.Print(fprintChanges(buys, sells))

	confirmed, err := confirmation("Do you want to automatically update portfolio info according to provided rebalance?")
	if err != nil {
		return "", fmt.Errorf("Fail to read user input")
	}
	if !confirmed {
		return "Portfolio wasn't automatically updated", nil
	}

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var msg strings.Builder
	if len(buys) > 0 {
		err = httpClient.AddStocks(ctx, buys)
		if err != nil {
			return "", fmt.Errorf("Failed to automatically update portfolio. (%w)", err)
		}
		msg.WriteString("Added successfully: ")
		for i := range buys {
			fmt.Fprintf(&msg, "%d of %q; ", buys[i].LotAmount, buys[i].Name)
		}
	}

	if len(sells) > 0 {
		for i := range sells {
			sells[i].LotAmount *= -1
		}
		err = httpClient.RemoveStocks(ctx, sells)
		if err != nil {
			fmt.Println(msg.String())
			return "", fmt.Errorf("Failed to remove stocks from portfolio. (%w)", err)
		}
	}

	return "Portfolio was successfully updated", nil
}

func sortRebalance(resp model.RebalanceResponse, originalStocks []model.Stock) (buys, sells []model.Stock) {
	for _, s := range resp.Stocks {
		for _, os := range originalStocks {
			if s.Code == os.Code {
				s.LotAmount = s.LotAmount - os.LotAmount
				break
			}
		}
		if s.LotAmount > 0 {
			buys = append(buys, s)
		} else if s.LotAmount < 0 {
			sells = append(sells, s)
		}
	}
	return
}

func fprintGeneralPortfolioInfo(resp model.RebalanceResponse) string {
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

	return msg.String()
}

func fprintChanges(buys, sells []model.Stock) string {
	var msg strings.Builder

	msg.WriteString("\n")

	if len(sells) > 0 {
		msg.WriteString("Sell:\n")
		for _, stock := range sells {
			fmt.Fprintf(
				&msg,
				"  - %s (%s): %d lot(s)\n",
				stock.Name,
				stock.Code,
				-stock.LotAmount,
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
				stock.LotAmount,
			)
		}
		msg.WriteString("\n")
	}

	return msg.String()
}
