package moex

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"stocks_calculator/internal/server/servererrors"
)

func couponTestHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	code := strings.TrimSuffix(path[strings.LastIndex(path, "/")+1:], ".json")
	if strings.HasPrefix(path, "/securities/") {
		switch code {
		case "UNKNOWN":
			fmt.Fprint(w, `{"description":{"columns":["name","value"],"data":[]}}`)
		case "SHARE":
			fmt.Fprint(w, `{"description":{"columns":["name","value"],"data":[["GROUP","stock_shares"]]}}`)
		case "USD":
			fmt.Fprint(w, `{"description":{"columns":["name","value"],"data":[["GROUP","stock_bonds"],["FACEUNIT","USD"]]}}`)
		default:
			fmt.Fprint(w, `{"description":{"columns":["name","value"],"data":[["GROUP","stock_bonds"],["FACEUNIT","SUR"]]}}`)
		}
		return
	}
	if !strings.HasPrefix(path, "/statistics/engines/stock/markets/bonds/bondization/") {
		http.Error(w, "unexpected path", http.StatusNotFound)
		return
	}
	switch code {
	case "USD":
		fmt.Fprint(w, `{"coupons":{"columns":["coupondate","value","faceunit"],"data":[["2030-01-01",14.75,"USD"],["2031-01-01",null,"USD"]]},"coupons.cursor":{"columns":["INDEX","TOTAL","PAGESIZE"],"data":[[0,2,20]]}}`)
	case "EMPTY":
		fmt.Fprint(w, `{"coupons":{"columns":["coupondate","value","faceunit"],"data":[]},"coupons.cursor":{"columns":["INDEX","TOTAL","PAGESIZE"],"data":[[0,0,20]]}}`)
	case "BROKEN":
		fmt.Fprint(w, `{"coupons":{"columns":["coupondate","value","faceunit"],"data":[]},"coupons.cursor":{"columns":["INDEX","TOTAL","PAGESIZE"],"data":[[0,1,20]]}}`)
	case "BADDATE":
		fmt.Fprint(w, `{"coupons":{"columns":["coupondate","value","faceunit"],"data":[["0000-00-00",5,"RUB"]]},"coupons.cursor":{"columns":["INDEX","TOTAL","PAGESIZE"],"data":[[0,1,20]]}}`)
	case "BADCURRENCY":
		fmt.Fprint(w, `{"coupons":{"columns":["coupondate","value","faceunit"],"data":[["2030-01-01",5,""]]},"coupons.cursor":{"columns":["INDEX","TOTAL","PAGESIZE"],"data":[[0,1,20]]}}`)
	case "BADAMOUNT":
		fmt.Fprint(w, `{"coupons":{"columns":["coupondate","value","faceunit"],"data":[["2030-01-01","unknown","RUB"]]},"coupons.cursor":{"columns":["INDEX","TOTAL","PAGESIZE"],"data":[[0,1,20]]}}`)
	case "CHANGED":
		if r.URL.Query().Get("start") == "0" {
			fmt.Fprint(w, `{"coupons":{"columns":["coupondate","value","faceunit"],"data":[["2029-01-01",5,"RUB"]]},"coupons.cursor":{"columns":["INDEX","TOTAL","PAGESIZE"],"data":[[0,2,1]]}}`)
		} else {
			fmt.Fprint(w, `{"coupons":{"columns":["coupondate","value","faceunit"],"data":[["2030-01-01",5,"RUB"]]},"coupons.cursor":{"columns":["INDEX","TOTAL","PAGESIZE"],"data":[[1,3,1]]}}`)
		}
	case "HTTPERR":
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
	default:
		if r.URL.Query().Get("start") == "0" {
			fmt.Fprint(w, `{"coupons":{"columns":["faceunit","value","coupondate"],"data":[["RUB",10.5,"2025-01-01"],["RUB",null,"2030-01-01"]]},"coupons.cursor":{"columns":["INDEX","TOTAL","PAGESIZE"],"data":[[0,3,2]]}}`)
		} else {
			fmt.Fprint(w, `{"coupons":{"columns":["coupondate","value","faceunit"],"data":[["2026-01-01",0,"SUR"]]},"coupons.cursor":{"columns":["INDEX","TOTAL","PAGESIZE"],"data":[[2,3,2]]}}`)
		}
	}
}

func TestGetBondCouponsPaginationAndAmounts(t *testing.T) {
	client := bondTestClient(t, couponTestHandler)
	result, err := client.GetBondCoupons(context.Background(), "BOND")
	if err != nil {
		t.Fatal(err)
	}
	if result.Code != "BOND" || len(result.Coupons) != 3 {
		t.Fatalf("unexpected response: %+v", result)
	}
	if result.Coupons[0].PaymentDate != "2025-01-01" || result.Coupons[0].Amount == nil || result.Coupons[0].Amount.Price != 10.5 || result.Coupons[0].Amount.Currency != "RUB" ||
		result.Coupons[1].PaymentDate != "2026-01-01" || result.Coupons[1].Amount == nil || result.Coupons[1].Amount.Price != 0 || result.Coupons[1].Amount.Currency != "RUB" ||
		result.Coupons[2].PaymentDate != "2030-01-01" || result.Coupons[2].Amount != nil {
		t.Fatalf("unexpected coupon values: %+v", result.Coupons)
	}
}

func TestGetBondCouponsEmptyAndErrors(t *testing.T) {
	client := bondTestClient(t, couponTestHandler)
	empty, err := client.GetBondCoupons(context.Background(), "EMPTY")
	if err != nil || empty.Coupons == nil || len(empty.Coupons) != 0 {
		t.Fatalf("empty schedule: %+v, %v", empty, err)
	}
	for _, tc := range []struct {
		code string
		want error
	}{
		{"UNKNOWN", servererrors.ErrMoexBondNotFound},
		{"SHARE", servererrors.ErrMoexBondNotFound},
		{"BROKEN", nil}, {"CHANGED", nil}, {"BADDATE", nil}, {"BADCURRENCY", nil}, {"BADAMOUNT", nil}, {"HTTPERR", nil},
	} {
		t.Run(tc.code, func(t *testing.T) {
			_, err := client.GetBondCoupons(context.Background(), tc.code)
			if err == nil || (tc.want != nil && !errors.Is(err, tc.want)) {
				t.Fatalf("got %v, want %v", err, tc.want)
			}
		})
	}
}

func TestForeignBondCoupons(t *testing.T) {
	client := bondTestClient(t, couponTestHandler)
	result, err := client.GetBondCoupons(context.Background(), "USD")
	if err != nil || len(result.Coupons) != 2 || result.Coupons[0].Amount == nil || result.Coupons[0].Amount.Currency != "USD" || result.Coupons[0].Amount.Price != 14.75 || result.Coupons[1].Amount != nil {
		t.Fatalf("foreign coupon schedule: %+v, %v", result, err)
	}
}
