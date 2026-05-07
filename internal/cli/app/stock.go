package app

import (
	"context"
	"fmt"
	"stocks_calculator/internal/cli/client"
	"stocks_calculator/internal/cli/config"
	"stocks_calculator/internal/model"
	"strconv"
	"strings"
	"time"
)

func StockInfo() (string, error) {
	httpClient := client.New(config.Url(), config.UserConfig.Token, config.UserConfig.RefToken)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stocks, err := httpClient.GetStocks(ctx)
	if err != nil {
		return "", fmt.Errorf("Failded to retrive info about portfolio. %w", err)
	}
	groups, err := httpClient.GetGroups(ctx)
	if err != nil {
		return "", fmt.Errorf("Failded to retrive info about portfolio. %w", err)
	}

	if len(stocks) == 0 {
		return "No stocks in portfolio. To add stocks use \"stock add\" command", nil
	}

	var builder strings.Builder
	for i := range groups {
		fmt.Fprintf(&builder, "Stocks in group %s (%d %%):\n", groups[i].Name, groups[i].Weight)
		ind := 1
		for j := range stocks {
			if stocks[j].GroupName == groups[i].Name {
				fmt.Fprintf(&builder, "%d)\n%s", ind, sprintStock(stocks[j]))
				ind++
			}
		}
		builder.WriteRune('\n')
	}

	return builder.String(), nil
}

func AddStocks(args []string) (string, error) {
	stocks, err := parseStockAgs(args, true)
	if err != nil {
		return "", err
	}

	httpClient := client.New(config.Url(), config.UserConfig.Token, config.UserConfig.RefToken)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	groups, err := httpClient.GetGroups(ctx)
	if err != nil {
		return "", fmt.Errorf("Failed to load group information. %w", err)
	}
	if len(groups) == 0 {
		return "", fmt.Errorf("To add stocks you need to add group first")
	}

	response, err := httpClient.FindStocks(ctx, stockNames(stocks))
	if err != nil {
		return "", fmt.Errorf("Failed to add stocks. %w", err)
	}

	stocksToAdd, err := confirmAdditions(groups, stocks, response)
	if err != nil {
		return "", err
	}
	if len(stocksToAdd) == 0 {
		return "No stocks were added", nil
	}

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = httpClient.AddStocks(ctx, stocksToAdd)
	if err != nil {
		return "", fmt.Errorf("Failed to add stocks. %w", err)
	}

	return fmt.Sprintf("%d/%d stock(s) added successfully", len(stocksToAdd), len(stocks)), nil
}

func parseStockAgs(args []string, isName bool) ([]model.Stock, error) {
	stocks := make([]model.Stock, 0, len(args))
	wasNum := false
	for i := range args {
		amount, err := strconv.Atoi(args[i])

		if err != nil {
			if isName {
				stocks = append(stocks, model.Stock{Name: args[i], LotAmount: 1})
			} else {
				stocks = append(stocks, model.Stock{Code: args[i], LotAmount: 1})
			}
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

		stocks[len(stocks)-1].LotAmount = amount
		wasNum = true
	}

	return stocks, nil
}

func confirmAdditions(
	groups []model.Group,
	stocks []model.Stock,
	foundStocks []model.FoundStock,
) ([]model.Stock, error) {
	stocksToAdd := []model.Stock{}

	for i := range foundStocks {
		if foundStocks[i].ErrMsg != "" {
			fmt.Printf("Failed to add stock %q: %s\n", stocks[i].Name, foundStocks[i].ErrMsg)
			continue
		}
		chosenStock, err := processFoundStock(foundStocks[i], stocks[i].Name, stocks[i].LotAmount, groups)
		if err != nil {
			fmt.Printf("Failed to add stock %q: %s\n", stocks[i].Name, err.Error())
			continue
		}
		if chosenStock == nil {
			continue
		}
		stocksToAdd = append(
			stocksToAdd,
			*chosenStock,
		)
	}

	return stocksToAdd, nil
}

func chooseFromFoundStock(foundStock model.FoundStock, query string) (int, error) {
	fmt.Print(saveCursorPos)
	fmt.Printf("\nSearching for %q. Found:\n", query)

	for i, s := range foundStock.SearchResults {
		fmt.Printf("\t%d) Name: %s, Code: %s, Price: %.2f (руб)\n", i+1, s.Name, s.Code, s.Price)
	}

	index, err := askNumber(0, len(foundStock.SearchResults), "Choose stock index. (0 if none corresponds to search query)")

	fmt.Print(restoreCursorPos)
	fmt.Print(clearUntilEnd)
	return index - 1, err
}

func processFoundStock(foundStock model.FoundStock, query string, amount int, groups []model.Group) (*model.Stock, error) {
	ind, err := chooseFromFoundStock(foundStock, query)
	if err != nil {
		return nil, err
	}
	if ind == -1 {
		return nil, nil
	}

	chosenStock := foundStock.SearchResults[ind]
	if chosenStock.LotAmount == 0 {
		groupInd, err := askGroup(groups)
		if err != nil {
			return nil, err
		}
		chosenStock.GroupName = groups[groupInd].Name
	}

	msg := buildConfirmationMsg(chosenStock, amount)
	confirmed, err := confirmation(msg)
	if err != nil {
		return nil, fmt.Errorf("Failed to read user input")
	}
	if !confirmed {
		return nil, nil
	}

	chosenStock.LotAmount = amount
	return &chosenStock, nil
}

func buildConfirmationMsg(chosenStock model.Stock, amount int) string {
	var msg string

	if chosenStock.LotAmount > 0 {
		msg = fmt.Sprintf(
			"Append %d lot(s) of %q to group %q (previously group contained %d lot(s))",
			amount,
			chosenStock.Name,
			chosenStock.GroupName,
			chosenStock.LotAmount,
		)
	} else {
		msg = fmt.Sprintf(
			"Add %d lot(s) of %q to group %q",
			amount,
			chosenStock.Name,
			chosenStock.GroupName,
		)
	}

	return msg
}

func askGroup(groups []model.Group) (int, error) {
	fmt.Print(saveCursorPos)
	fmt.Println("Please choose group to add stocks to. List of available groups:")
	for i := range groups {
		fmt.Printf("%d) %s\n", i+1, groups[i].Name)
	}

	index, err := askNumber(1, len(groups), "Choose group index")
	if err != nil {
		return -1, err
	}

	fmt.Print(restoreCursorPos)
	fmt.Print(clearUntilEnd)
	return index - 1, nil
}

type (
	RemoveStockOptions struct {
		All bool
	}
)

func RemoveStock(options RemoveStockOptions, args []string) (string, error) {
	stocks, err := parseStockAgs(args, false)
	if err != nil {
		return "", err
	}

	httpClient := client.New(config.Url(), config.UserConfig.Token, config.UserConfig.RefToken)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	currentStocks, err := httpClient.GetStocks(ctx)
	if err != nil {
		return "", fmt.Errorf("Failed to get current stock information. %w", err)
	}

	toRemove, err := confirmRemovals(stocks, currentStocks, options.All)
	if err != nil {
		return "", err
	}

	if len(toRemove) == 0 {
		return "Nothing was removed", nil
	}

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = httpClient.RemoveStocks(ctx, toRemove)
	if err != nil {
		return "", fmt.Errorf("Failed to remove stocks. (%w)", err)
	}

	return fmt.Sprintf("%d/%d removed successfully", len(toRemove), len(stocks)), nil
}

func confirmRemovals(stocksToRemove []model.Stock, currentStocks []model.Stock, all bool) ([]model.Stock, error) {
	toRemove := make([]model.Stock, 0, len(stocksToRemove))

outer:
	for _, toRem := range stocksToRemove {
		for _, curr := range currentStocks {
			if toRem.Code != curr.Code {
				continue
			}

			amount := toRem.LotAmount
			if amount > curr.LotAmount || all {
				amount = curr.LotAmount
			}
			confirmed, err := confirmation(fmt.Sprintf(
				"Remove %d/%d lot(s) of %s from group %s",
				amount,
				curr.LotAmount,
				curr.Name,
				curr.GroupName,
			))
			if err != nil {
				return nil, fmt.Errorf("Failed to read user input")
			}
			if !confirmed {
				continue outer
			}

			curr.LotAmount = amount
			toRemove = append(toRemove, curr)
			continue outer
		}

		fmt.Printf("Failed to remove %s. No such stock in portfolio", toRem.Code)
	}

	return toRemove, nil
}

func stockNames(stocks []model.Stock) []string {
	names := make([]string, len(stocks))
	for i, s := range stocks {
		names[i] = s.Name
	}
	return names
}

func sprintStock(stock model.Stock) string {
	groupLine := "  Group: " + stock.GroupName + "\n"
	return fmt.Sprintf(
		"  Name: %s\n  Code: %s\n  Current amount of lots: %d\n%s  Current price: %.4f(RUB)\n  Lot size: %d share(s)\n  Total value: %.4f(RUB)\n",
		stock.Name, stock.Code, stock.LotAmount, groupLine, stock.Price, stock.LotSize, stock.Price*float64(stock.LotAmount)*float64(stock.LotSize),
	)
}
