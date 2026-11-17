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
	rates := map[string]float64{"RUB": 1}
	needsRates := false
	for _, bond := range bonds {
		for _, currency := range []string{bond.CleanPrice.Currency, bond.AccruedInterest.Currency} {
			if !supportedCurrency(currency) {
				return model.BondRebalanceResponse{}, servererrors.ErrUnsupportedCurrency
			}
			needsRates = needsRates || currency != "RUB"
		}
	}
	if needsRates {
		rates, err = a.moex.GetCurrencyRates(ctx)
		if err != nil {
			return model.BondRebalanceResponse{}, err
		}
	}
	for i, bond := range bonds {
		cleanRate, cleanOK := rates[bond.CleanPrice.Currency]
		accruedRate, accruedOK := rates[bond.AccruedInterest.Currency]
		if !cleanOK || !accruedOK || cleanRate <= 0 || accruedRate <= 0 || math.IsNaN(cleanRate) || math.IsNaN(accruedRate) || math.IsInf(cleanRate, 0) || math.IsInf(accruedRate, 0) {
			return model.BondRebalanceResponse{}, servererrors.ErrMoexUnhandled
		}
		price := bond.CleanPrice.Price*cleanRate + bond.AccruedInterest.Price*accruedRate
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
