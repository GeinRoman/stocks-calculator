package app

import (
	"context"
	"testing"

	"stocks_calculator/internal/model"
)

type couponTestMoex struct {
	MoexApi
	code string
}

func (m *couponTestMoex) GetBondCoupons(_ context.Context, code string) (model.BondCouponsResponse, error) {
	m.code = code
	return model.BondCouponsResponse{Code: code, Coupons: []model.Coupon{}}, nil
}

func TestGetBondCouponsDoesNotRequireHoldings(t *testing.T) {
	moex := &couponTestMoex{}
	a := New(nil, nil, moex)
	result, err := a.GetBondCoupons(context.Background(), "ABC")
	if err != nil || moex.code != "ABC" || result.Code != "ABC" || result.Coupons == nil {
		t.Fatalf("response %+v, code %q, error %v", result, moex.code, err)
	}
}
