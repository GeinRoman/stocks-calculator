package app

import (
	"fmt"
	"stocks_calculator/internal/cli/client"
	"strconv"
	"strings"
)

func StockInfo() (string, error) {
	info, err := client.GetStocksInfo()
	if err != nil {
		return "", fmt.Errorf("Failded to retrive info about portfolio")
	}

	if len(info.Stocks) == 0 {
		return "No stocks in portfolio. To add stocks use \"stock add\" command", nil
	}

	var builder strings.Builder
	for i := range info.Groups {
		fmt.Fprintf(&builder, "Stocks in group %s (%d %%):\n", info.Groups[i].Name, info.Groups[i].Weight)
		ind := 1
		for j := range info.Stocks {
			if info.Stocks[j].GroupId == uint(i+1) {
				fmt.Fprintf(
					&builder, "%d)\n%s",
					ind, sprintStock(info.Stocks[j], ""),
				)
				ind++
			}
		}
		builder.WriteRune('\n')
	}

	return builder.String(), nil
}

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
	stocks, err := parseStockAgs(args)
	if err != nil {
		return "", err
	}

	groups, err := client.GetGroups()
	if err != nil {
		return "", fmt.Errorf("Failed to load group information")
	}
	if len(groups) == 0 {
		return "", fmt.Errorf("To add stocks you need to add group first")
	}

	response, err := client.FindStock(stockNames(stocks))
	if err != nil {
		return "", fmt.Errorf("Failed to add stocks")
	}

	stocksToAdd, err := confirmAdditions(groups, stocks, response)
	if err != nil {
		return "", err
	}
	if len(stocksToAdd) == 0 {
		return "No stocks were added", nil
	}

	err = client.AddStock(stocksToAdd)
	if err != nil {
		return "", fmt.Errorf("Failed to add stocks")
	}

	return fmt.Sprintf("%d/%d stock(s) added successfully", len(stocksToAdd), len(stocks)), nil
}

func parseStockAgs(args []string) ([]stock, error) {
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

func confirmAdditions(
	groups []client.Group,
	stocks []stock,
	response []client.StockInfo,
) (
	[]client.StockInfoShort,
	error,
) {
	stocksToAdd := []client.StockInfoShort{}

	for i := range response {
		if response[i].ErrorMsg != "" {
			fmt.Printf("Failed to add stock %q: %s\n", stocks[i].Name, response[i].ErrorMsg)
			continue
		}

		fmt.Printf("\nSearching for %q. Found:\n", stocks[i].Name)

		fmt.Print(sprintStock(response[i], ""))

		groupId := response[i].GroupId
		if groupId == 0 {
			var err error
			groupId, err = askGroup(groups)
			if err != nil {
				return nil, err
			}
		}

		msg := buildConfirmationMsg(response[i], groups[groupId-1].Name, stocks[i].Amount)
		confirmed, err := confirmation(msg)
		if err != nil {
			return nil, fmt.Errorf("Failed to read user input")
		}
		if !confirmed {
			continue
		}

		stocksToAdd = append(
			stocksToAdd,
			client.StockInfoShort{Name: response[i].Name, Code: response[i].Code, GroupId: groupId, Amount: stocks[i].Amount},
		)
	}

	return stocksToAdd, nil
}

func buildConfirmationMsg(foundStock client.StockInfo, group string, amount int) string {
	var msg string

	if foundStock.CurrentAmount > 0 {
		msg = fmt.Sprintf(
			"Append %d lot(s) of %q to group %q (previously group contained %d lot(s))",
			amount,
			foundStock.Name,
			group,
			foundStock.CurrentAmount,
		)
	} else {
		msg = fmt.Sprintf(
			"Add %d lot(s) of %q to group %q",
			amount,
			foundStock.Name,
			group,
		)
	}

	return msg
}

func askGroup(groups []client.Group) (uint, error) {
	fmt.Print(saveCursorPos)
	fmt.Println("Please choose group to add stocks to. List of available groups:")
	for i := range groups {
		fmt.Printf("%d) %s\n", i+1, groups[i].Name)
	}

	index, err := askNumber(1, len(groups), "Choose group index")
	if err != nil {
		return 0, err
	}

	fmt.Print(restoreCursorPos)
	fmt.Print(clearUntilEnd)
	return uint(index), nil
}

type (
	RemoveStockOptions struct {
		All bool
	}
)

func RemoveStock(options RemoveStockOptions, args []string) (string, error) {
	stocks, err := parseStockAgs(args)
	if err != nil {
		return "", err
	}

	groups, err := client.GetGroups()
	if err != nil {
		return "", fmt.Errorf("Failed to load group information")
	}

	response, err := client.FindStock(stockNames(stocks))
	if err != nil {
		return "", fmt.Errorf("Failed to remove stocks. (%w)", err)
	}

	toRemove, err := confirmRemovals(response, stocks, groups, options.All)
	if err != nil {
		return "", err
	}

	err = client.RemoveStock(toRemove)
	if err != nil {
		return "", fmt.Errorf("Failed to remove stocks. (%w)", err)
	}

	return fmt.Sprintf("%d/%d removed successfully", len(toRemove), len(stocks)), nil
}

func confirmRemovals(response []client.StockInfo, stocks []stock, groups []client.Group, all bool) ([]client.StockInfoShort, error) {
	toRemove := []client.StockInfoShort{}

	for i := range response {
		if response[i].ErrorMsg != "" {
			fmt.Printf("Failed to remove stock %q: %s\n", stocks[i].Name, response[i].ErrorMsg)
			continue
		}

		if response[i].GroupId == 0 {
			fmt.Printf("Failed to remove stock %q: not found in portfolio\n", stocks[i].Name)
			continue
		}

		fmt.Printf("\nSearching for %q. Found:\n", stocks[i].Name)

		group := groups[response[i].GroupId-1].Name
		fmt.Print(sprintStock(response[i], group))

		amount := stocks[i].Amount
		if amount > response[i].CurrentAmount || all {
			amount = response[i].CurrentAmount
		}

		confirmed, err := confirmation(fmt.Sprintf("Remove %d/%d of %s from group %s", amount, response[i].CurrentAmount, response[i].Name, group))
		if err != nil {
			return nil, fmt.Errorf("Failed to read user input")
		}
		if !confirmed {
			continue
		}

		toRemove = append(
			toRemove,
			client.StockInfoShort{Name: response[i].Name, Code: response[i].Code, GroupId: response[i].GroupId, Amount: stocks[i].Amount},
		)
	}

	return toRemove, nil
}

func stockNames(stocks []stock) []string {
	names := make([]string, len(stocks))
	for i, s := range stocks {
		names[i] = s.Name
	}
	return names
}

func sprintStock(stock client.StockInfo, group string) string {
	groupLine := ""
	if group != "" {
		groupLine = "  Group: " + group + "\n"
	}
	return fmt.Sprintf(
		"  Name: %s\n  Code: %s\n  Current amount of lots: %d\n%s  Current price: %.4f\n  Lot size: %d share(s)\n",
		stock.Name, stock.Code, stock.CurrentAmount, groupLine, stock.CurrentPrice, stock.LotSize,
	)
}
