package app

import (
	"context"
	"errors"
	"testing"

	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/servererrors"
)

type rebalanceBondRepo struct {
	Repository
	bonds  []model.Bond
	groups []model.BondGroup
	err    error
}

func (r rebalanceBondRepo) GetBonds(context.Context, int) ([]model.Bond, error) {
	return append([]model.Bond(nil), r.bonds...), r.err
}

func (r rebalanceBondRepo) GetBondGroups(context.Context, int) ([]model.BondGroup, error) {
	return r.groups, r.err
}

func TestRebalanceBondsDirtyPriceAndLots(t *testing.T) {
	payment := 1000.0
	repo := rebalanceBondRepo{
		bonds:  []model.Bond{{Name: "Bond", Code: "A", GroupName: "G", LotAmount: 3, LotSize: 2}},
		groups: []model.BondGroup{{Name: "G", Weight: 100}},
	}
	moex := bondTestMoex{info: []model.MoexBondInfoResult{{Bond: model.Bond{CleanPrice: 10, AccruedInterest: 2, LotSize: 2, MaturityDate: "2030-01-01", MaturityPayment: &payment}}}}
	a := New(repo, nil, moex)
	for _, noSell := range []bool{false, true} {
		response, err := a.RebalanceBonds(context.Background(), 1, noSell, 25)
		if err != nil {
			t.Fatal(err)
		}
		if response.TotalPrevCost != 72 || response.CostDiff != 24 || len(response.Bonds) != 1 || response.Bonds[0].LotAmount != 4 {
			t.Fatalf("noSell=%v: unexpected proposal %+v", noSell, response)
		}
		bond := response.Bonds[0]
		if bond.Name != "Bond" || bond.CleanPrice != 10 || bond.AccruedInterest != 2 || bond.MaturityDate != "2030-01-01" || bond.MaturityPayment == nil || *bond.MaturityPayment != 1000 {
			t.Fatalf("bond metadata changed: %+v", bond)
		}
	}
}

func TestRebalanceBondsPreconditions(t *testing.T) {
	for _, tc := range []struct {
		name   string
		repo   rebalanceBondRepo
		info   []model.MoexBondInfoResult
		wanted error
	}{
		{name: "no bonds", wanted: servererrors.ErrBondsNotFound},
		{name: "no groups", repo: rebalanceBondRepo{bonds: []model.Bond{{Code: "A"}}}, info: []model.MoexBondInfoResult{{}}, wanted: servererrors.ErrGroupsNotFound},
		{name: "weights", repo: rebalanceBondRepo{bonds: []model.Bond{{Code: "A"}}, groups: []model.BondGroup{{Name: "G", Weight: 50}}}, info: []model.MoexBondInfoResult{{}}, wanted: servererrors.ErrGroupsAreNotWeighted},
		{name: "moex failure", repo: rebalanceBondRepo{bonds: []model.Bond{{Code: "A"}}}, info: []model.MoexBondInfoResult{{Err: errors.New("unavailable")}}, wanted: servererrors.ErrMoexUnhandled},
		{name: "repository failure", repo: rebalanceBondRepo{err: servererrors.ErrNoDefaultProfile}, wanted: servererrors.ErrNoDefaultProfile},
		{name: "invalid price", repo: rebalanceBondRepo{bonds: []model.Bond{{Code: "A"}}, groups: []model.BondGroup{{Name: "G", Weight: 100}}}, info: []model.MoexBondInfoResult{{Bond: model.Bond{LotSize: 1}}}, wanted: servererrors.ErrMoexUnhandled},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := New(tc.repo, nil, bondTestMoex{info: tc.info})
			_, err := a.RebalanceBonds(context.Background(), 1, false, 0)
			if !errors.Is(err, tc.wanted) {
				t.Fatalf("got %v, want %v", err, tc.wanted)
			}
		})
	}
}

func TestRebalanceBondsIndependentGroupsAndNoSell(t *testing.T) {
	repo := rebalanceBondRepo{
		bonds:  []model.Bond{{Code: "A", GroupName: "First", LotAmount: 3}, {Code: "B", GroupName: "Second", LotAmount: 1}},
		groups: []model.BondGroup{{Name: "First", Weight: 50}, {Name: "Second", Weight: 50}},
	}
	moex := bondTestMoex{info: []model.MoexBondInfoResult{
		{Bond: model.Bond{CleanPrice: 90, AccruedInterest: 10, LotSize: 1}},
		{Bond: model.Bond{CleanPrice: 80, AccruedInterest: 20, LotSize: 1}},
	}}
	a := New(repo, nil, moex)
	for _, tc := range []struct {
		noSell bool
		diff   float64
		first  int
		second int
	}{
		{noSell: false, first: 2, second: 2},
		{noSell: true, diff: 200, first: 3, second: 3},
	} {
		result, err := a.RebalanceBonds(context.Background(), 1, tc.noSell, tc.diff)
		if err != nil || result.TotalPrevCost != 400 || result.CostDiff != tc.diff || len(result.Bonds) != 2 || result.Bonds[0].LotAmount != tc.first || result.Bonds[1].LotAmount != tc.second {
			t.Fatalf("noSell=%v: result %+v, error %v", tc.noSell, result, err)
		}
	}
}
