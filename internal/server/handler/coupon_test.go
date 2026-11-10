package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/servererrors"
)

type couponTestApp struct {
	App
	code string
	err  error
}

func (a *couponTestApp) GetBondCoupons(_ context.Context, code string) (model.BondCouponsResponse, error) {
	a.code = code
	return model.BondCouponsResponse{Code: code, Coupons: []model.Coupon{{PaymentDate: "2030-01-01"}}}, a.err
}

func TestBondCouponsEndpoint(t *testing.T) {
	a := &couponTestApp{}
	h := NewHttpServer(a, bondTestTokens{}, 0).Handler
	if w := bondRequest(h, http.MethodGet, "/bondcoupons?code=ABC", "", ""); w.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated status: %d", w.Code)
	}
	if w := bondRequest(h, http.MethodGet, "/bondcoupons", "", "alice"); w.Code != http.StatusBadRequest {
		t.Fatalf("missing code status: %d", w.Code)
	}
	w := bondRequest(h, http.MethodGet, "/bondcoupons?code=ABC", "", "alice")
	var got model.BondCouponsResponse
	if w.Code != http.StatusOK || json.Unmarshal(w.Body.Bytes(), &got) != nil || a.code != "ABC" || got.Code != "ABC" || len(got.Coupons) != 1 || got.Coupons[0].Amount != nil {
		t.Fatalf("status %d, response %s", w.Code, w.Body.String())
	}
	for _, tc := range []struct {
		err    error
		status int
	}{
		{servererrors.ErrMoexBondNotFound, http.StatusNotFound},
		{servererrors.ErrUnsupportedCurrency, http.StatusUnprocessableEntity},
		{context.DeadlineExceeded, http.StatusGatewayTimeout},
	} {
		a.err = tc.err
		if w := bondRequest(h, http.MethodGet, "/bondcoupons?code=ABC", "", "alice"); w.Code != tc.status {
			t.Errorf("error %v: got %d, want %d", tc.err, w.Code, tc.status)
		}
	}
}
