package app

import (
	"context"
	"errors"
	"fmt"
	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/servererrors"
)

func (a *app) AddStocks(ctx context.Context, userId int, stocks []model.Stock) error {
	curStocks, err := a.repo.GetStocks(ctx, userId)
	if err != nil {
		return err
	}

	var (
		toUpdate, toInsert []model.Stock
		amountDiff         []int
	)

outer:
	for i := range stocks {
		if stocks[i].Amount <= 0 {
			return fmt.Errorf("%w: %q", servererrors.ErrWrongStockAmount, stocks[i].Name)
		}
		for j := range curStocks {
			if stocks[i].Code == curStocks[j].Code {
				if stocks[i].GroupName != curStocks[j].GroupName {
					return fmt.Errorf("Failed to add stocks. %w: %q in %q", servererrors.ErrWrongStockGroup, stocks[i].Name, curStocks[j].GroupName)
				}

				updatedAmount := curStocks[j].Amount + stocks[i].Amount
				amountDiff = append(amountDiff, stocks[i].Amount)
				stocks[i].Amount = updatedAmount
				toUpdate = append(toUpdate, stocks[i])
				continue outer
			}
		}

		toInsert = append(toInsert, stocks[i])
	}

	err = a.repo.InsertStocks(ctx, userId, toInsert)
	if err != nil {
		return err
	}
	err = a.repo.UpdateStocksAmount(ctx, userId, toUpdate, amountDiff)

	return err
}

func (a *app) RemoveStocks(ctx context.Context, userId int, stocks []model.Stock) error {
	curStocks, err := a.repo.GetStocks(ctx, userId)
	if err != nil {
		return err
	}

	var (
		toUpdate, toRemove []model.Stock
		amountDiff         []int
	)
outer:
	for i := range stocks {
		for j := range curStocks {
			if stocks[i].Code == curStocks[j].Code {
				if stocks[i].GroupName != curStocks[j].GroupName {
					return fmt.Errorf("Failed to remove stocks. %w: %q in %q", servererrors.ErrWrongStockGroup, stocks[i].Name, curStocks[j].GroupName)
				}
				newAmount := curStocks[j].Amount - stocks[i].Amount
				if newAmount <= 0 {
					stocks[i].Amount = curStocks[j].Amount
					toRemove = append(toRemove, stocks[i])
				} else {
					amountDiff = append(amountDiff, -stocks[i].Amount)
					stocks[i].Amount = newAmount
					toUpdate = append(toUpdate, stocks[i])
				}
				continue outer
			}
		}
		return fmt.Errorf("Failed to remove. %w: %q", servererrors.ErrStockNotFound, stocks[i].Name)
	}

	err = a.repo.UpdateStocksAmount(ctx, userId, toUpdate, amountDiff)
	if err != nil {
		return err
	}

	return a.repo.RemoveStocks(ctx, userId, toRemove)
}

func (a *app) GetStocks(ctx context.Context, userId int) ([]model.Stock, error) {
	stocks, err := a.repo.GetStocks(ctx, userId)
	if err != nil {
		return nil, err
	}
	codes := make([]string, len(stocks))
	for i := range stocks {
		codes[i] = stocks[i].Code
	}
	prices := a.moex.GetPrices(ctx, codes)
	for i := range prices {
		if prices[i].Err == nil {
			stocks[i].Price = prices[i].Price
		}
	}

	return stocks, nil
}

func (a *app) FindStocks(ctx context.Context, names []string) []model.FoundStock {
	result := a.moex.FindStocks(ctx, names)
	stocks := make([]model.FoundStock, len(result))
	for i := range result {
		switch {
		case errors.Is(result[i].Err, servererrors.ErrMoexStockNotFound):
			stocks[i].ErrMsg = fmt.Sprintf("%s: %q", servererrors.ErrMoexStockNotFound.Error(), names[i])
		case result[i].Err != nil:
			stocks[i].ErrMsg = fmt.Sprintf("%s", servererrors.ErrMoexUnhandled.Error())
		default:
			stocks[i].Stock = result[i].Stock
		}
	}
	return stocks
}
