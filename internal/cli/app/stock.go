package app

import (
	"fmt"
	"stocks_calculator/internal/cli/client"
	"strconv"
)

type (
	AddStockOptions struct {
		Group string
	}
	stock struct {
		Name   string
		Amount int
	}
)

func AddStock(options AddStockOptions, args []string) (string, error) {
	stocks, err := parseAddStocksAgs(args)
	if err != nil {
		return "", err
	}

	if options.Group == "" {
		options.Group, err = askGroup()
		if err != nil {
			return "", err
		}
	}

	response, err := client.FindStock(client.FindStockBody{
		Stocks: stockNames(stocks),
		Group:  options.Group,
	})
	if err != nil {
		return "", fmt.Errorf("Failed to add stocks. (%w)", err)
	}
	group := response.GroupFound

	addStocksBody, err := proccessFindStocksResponse(group, stocks, response.FoundStocks)
	if err != nil {
		return "", err
	}
	if len(addStocksBody.Stocks) == 0 {
		return "No stocks were added", nil
	}

	err = client.AddStock(addStocksBody)
	if err != nil {
		return "", fmt.Errorf("Failed to add stocks.")
	}

	return fmt.Sprintf("%d/%d added successfully", len(addStocksBody.Stocks), len(stocks)), nil
}

func parseAddStocksAgs(args []string) ([]stock, error) {
	stocks := make([]stock, 0, len(args))
	wasNum := false
	for i := range args {
		amount, err := strconv.Atoi(args[i])

		if err != nil {
			stocks = append(stocks, stock{args[i], 1})
			wasNum = false
			continue
		}

		if amount <= 0 {
			return nil, fmt.Errorf("lot amount must be greater than zero, got \"%s %s\"", args[i-1], args[i])
		}

		if i == 0 {
			return nil, fmt.Errorf("Expected stock name, got number %q", args[i])
		}

		if wasNum {
			return nil, fmt.Errorf("Expected stock name after %q, got number %q", args[i-1], args[i])
		}

		stocks[len(stocks)-1].Amount = amount
		wasNum = true
	}

	return stocks, nil
}

func proccessFindStocksResponse(
	group string,
	stocks []stock,
	response []client.FoundStock,
) (client.AddStockBody, error) {
	addStocksBody := client.AddStockBody{
		Stocks: []client.StockBody{},
		Group:  group,
	}

	for i := range response {
		fmt.Printf("\nSearching for %q. Found:\n", stocks[i].Name)

		if response[i].ErrorMsg != "" {
			fmt.Printf("Failed to add stock: %s\n", response[i].ErrorMsg)
			continue
		}

		msg := buildConfirmationMsg(response[i], group, stocks[i].Amount)
		confirmed, err := confirmation(msg)
		if err != nil {
			return addStocksBody, fmt.Errorf("Failed to read user input")
		}
		if !confirmed {
			continue
		}

		addStocksBody.Stocks = append(
			addStocksBody.Stocks,
			client.StockBody{Name: response[i].Name, Code: response[i].Code, Amount: stocks[i].Amount},
		)
	}

	return addStocksBody, nil
}

func buildConfirmationMsg(foundStock client.FoundStock, group string, amount int) string {
	var actionMsg string

	if foundStock.PrevAmount > 0 {
		actionMsg = fmt.Sprintf(
			"Append %d lot(s) of %q to group %q (previously group contained %d lot(s))",
			amount,
			foundStock.Name,
			group,
			foundStock.PrevAmount,
		)
	} else {
		actionMsg = fmt.Sprintf(
			"Add %d lot(s) of %q to group %q",
			amount,
			foundStock.Name,
			group,
		)
	}

	return fmt.Sprintf(
		`  Name: %s
  Code: %s
  Price: %.2f
  Lot size: %d shares.
%s?`,
		foundStock.Name,
		foundStock.Code,
		foundStock.CurrentPrice,
		foundStock.LotSize,
		actionMsg,
	)

}

func askGroup() (string, error) {
	groups, err := client.GetGroups()
	if err != nil {
		return "", fmt.Errorf("Failed to load group information")
	}
	if len(groups) == 0 {
		return "", fmt.Errorf("To add stocks you need to add group first.")
	}
	fmt.Println("Please choose group to add stocks to. List of available groups:")
	for i := range groups {
		fmt.Printf("%d) %s\n", i+1, groups[i].Name)
	}

	index, err := askNumber(1, len(groups), "Choose group index")
	if err != nil {
		return "", err
	}
	return groups[index-1].Name, nil
}

func stockNames(stocks []stock) []string {
	names := make([]string, len(stocks))
	for i, s := range stocks {
		names[i] = s.Name
	}
	return names
}
