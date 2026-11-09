package handler

import (
	"context"
	"net/http"
	"testing"

	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/servererrors"
)

type rebalanceBondApp struct {
	App
	user   int
	noSell bool
	diff   float64
	err    error
}

func (a *rebalanceBondApp) RebalanceBonds(_ context.Context, user int, noSell bool, diff float64) (model.BondRebalanceResponse, error) {
	a.user, a.noSell, a.diff = user, noSell, diff
	return model.BondRebalanceResponse{Bonds: []model.Bond{{Code: "A", LotAmount: 2}}, TotalPrevCost: 100, CostDiff: 50}, a.err
}

func TestRebalanceBondsRoute(t *testing.T) {
	app := &rebalanceBondApp{}
	h := NewHttpServer(app, bondTestTokens{}, 0).Handler
	if w := bondRequest(h, http.MethodGet, "/rebalancebonds?nosell=true&valdiff=50", "", ""); w.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized status = %d", w.Code)
	}
	for _, query := range []string{"", "?nosell=true", "?nosell=bad&valdiff=1", "?nosell=true&valdiff=0"} {
		if w := bondRequest(h, http.MethodGet, "/rebalancebonds"+query, "", "alice"); w.Code != http.StatusBadRequest {
			t.Fatalf("%s: status = %d", query, w.Code)
		}
	}
	w := bondRequest(h, http.MethodGet, "/rebalancebonds?nosell=true&valdiff=50", "", "alice")
	if w.Code != http.StatusOK || app.user != 11 || !app.noSell || app.diff != 50 {
		t.Fatalf("status %d, application %+v", w.Code, app)
	}
	app.err = servererrors.ErrBondsNotFound
	if w := bondRequest(h, http.MethodGet, "/rebalancebonds?nosell=false&valdiff=0", "", "alice"); w.Code != http.StatusPreconditionRequired {
		t.Fatalf("no bonds status = %d", w.Code)
	}
}
