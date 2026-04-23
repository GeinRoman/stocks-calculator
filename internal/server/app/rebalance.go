package app

import (
	"context"
	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/servererrors"
)

func (a *app) Rebalance(
	ctx context.Context,
	userId int,
	noSell bool,
	desiredDiff float64,
) (model.RebalanceResponse, error) {
	stocks, err := a.GetStocks(ctx, userId)
	if err != nil {
		return model.RebalanceResponse{}, err
	}
	if len(stocks) == 0 {
		return model.RebalanceResponse{}, servererrors.ErrStocksNotFound
	}

	groups, err := a.repo.GetGroups(ctx, userId)
	if err != nil {
		return model.RebalanceResponse{}, err
	}
	if len(groups) == 0 {
		return model.RebalanceResponse{}, servererrors.ErrGroupsNotFound
	}
	weight := 0
	for _, g := range groups {
		weight += g.Weight
	}
	if weight != 100 {
		return model.RebalanceResponse{}, servererrors.ErrGroupsAreNotWeighted
	}

	helper := newRebalanceHelper(stocks, groups, desiredDiff, noSell)
	newAmounts := helper.computeAmounts()
	totalEndCost := helper.getStocksCost(stocks, newAmounts)
	newAmounts = helper.minimizeGapAcrossAllGroups(
		newAmounts,
		stocks,
		desiredDiff,
		totalEndCost-helper.totalPriorCost,
	)
	stocks, totalEndCost = helper.applyAmounts(stocks, newAmounts)

	return model.RebalanceResponse{
		Stocks:        stocks,
		TotalPrevCost: helper.totalPriorCost,
		CostDiff:      totalEndCost - helper.totalPriorCost,
	}, nil
}
