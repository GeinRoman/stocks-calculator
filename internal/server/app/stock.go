package app

import (
	"context"
	"fmt"
	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/servererrors"
)

func (a *app) AddStocks(ctx context.Context, userId int, stocks []model.Stock) error {
	curStocks, err := a.repo.GetStocks(ctx, userId)
	if err != nil {
		return err
	}

	var toUpdate, toInsert []model.Stock

outer:
	for i := range stocks {
		if stocks[i].LotAmount <= 0 {
			return fmt.Errorf("%w: %q", servererrors.ErrWrongStockAmount, stocks[i].Name)
		}
		for j := range curStocks {
			if stocks[i].Code == curStocks[j].Code {
				if stocks[i].GroupName != curStocks[j].GroupName {
					return fmt.Errorf("Failed to add stocks. %w: %q in %q", servererrors.ErrWrongStockGroup, stocks[i].Name, curStocks[j].GroupName)
				}

				updatedAmount := curStocks[j].LotAmount + stocks[i].LotAmount
				stocks[i].LotAmount = updatedAmount
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
	err = a.repo.UpdateStocksAmount(ctx, userId, toUpdate)

	return err
}

func (a *app) RemoveStocks(ctx context.Context, userId int, stocks []model.Stock) error {
	curStocks, err := a.repo.GetStocks(ctx, userId)
	if err != nil {
		return err
	}

	var toUpdate, toRemove []model.Stock
outer:
	for i := range stocks {
		for j := range curStocks {
			if stocks[i].Code == curStocks[j].Code {
				if stocks[i].GroupName != curStocks[j].GroupName {
					return fmt.Errorf("Failed to remove stocks. %w: %q in %q", servererrors.ErrWrongStockGroup, stocks[i].Name, curStocks[j].GroupName)
				}
				newAmount := curStocks[j].LotAmount - stocks[i].LotAmount
				if newAmount <= 0 {
					toRemove = append(toRemove, stocks[i])
				} else {
					stocks[i].LotAmount = newAmount
					toUpdate = append(toUpdate, stocks[i])
				}
				continue outer
			}
		}
		return fmt.Errorf("Failed to remove. %w: %q", servererrors.ErrStockNotFound, stocks[i].Name)
	}

	err = a.repo.UpdateStocksAmount(ctx, userId, toUpdate)
	if err != nil {
		return err
	}

	return a.repo.RemoveStocks(ctx, userId, toRemove)
}

func (a *app) GetStocks(ctx context.Context, userId int) ([]model.Stock, error) {
	return a.repo.GetStocks(ctx, userId)
}
