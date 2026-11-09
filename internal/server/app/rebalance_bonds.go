package app

import (
	"context"
	"math"
	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/servererrors"
)

func (a *app) RebalanceBonds(ctx context.Context, userId int, noSell bool, desiredDiff float64) (model.BondRebalanceResponse, error) {
	bonds, err := a.GetBonds(ctx, userId)
	if err != nil {
		return model.BondRebalanceResponse{}, err
	}
	if len(bonds) == 0 {
		return model.BondRebalanceResponse{}, servererrors.ErrBondsNotFound
	}

	groups, err := a.repo.GetBondGroups(ctx, userId)
	if err != nil {
		return model.BondRebalanceResponse{}, err
	}
	if len(groups) == 0 {
		return model.BondRebalanceResponse{}, servererrors.ErrGroupsNotFound
	}
	stockGroups := make([]model.StockGroup, len(groups))
	weight := 0
	for i, group := range groups {
		weight += group.Weight
		stockGroups[i] = model.StockGroup{Name: group.Name, Weight: group.Weight}
	}
	if weight != 100 {
		return model.BondRebalanceResponse{}, servererrors.ErrGroupsAreNotWeighted
	}

	stocks := make([]model.Stock, len(bonds))
	for i, bond := range bonds {
		price := bond.CleanPrice + bond.AccruedInterest
		if price <= 0 || math.IsNaN(price) || math.IsInf(price, 0) || bond.LotSize <= 0 {
			return model.BondRebalanceResponse{}, servererrors.ErrMoexUnhandled
		}
		stocks[i] = model.Stock{Code: bond.Code, GroupName: bond.GroupName, Price: price, LotAmount: bond.LotAmount, LotSize: bond.LotSize}
	}

	helper := newRebalanceHelper(stocks, stockGroups, desiredDiff, noSell)
	amounts := helper.computeAmounts()
	endCost := helper.getStocksCost(stocks, amounts)
	amounts = helper.minimizeGapAcrossAllGroups(amounts, stocks, desiredDiff, endCost-helper.totalPriorCost)
	for i := range bonds {
		bonds[i].LotAmount = amounts[bonds[i].Code]
	}
	endCost = helper.getStocksCost(stocks, amounts)
	return model.BondRebalanceResponse{Bonds: bonds, TotalPrevCost: helper.totalPriorCost, CostDiff: endCost - helper.totalPriorCost}, nil
}
