package app

import (
	"context"
	"stocks_calculator/internal/model"
)

func (a *app) GetBondCoupons(ctx context.Context, code string) (model.BondCouponsResponse, error) {
	return a.moex.GetBondCoupons(ctx, code)
}
