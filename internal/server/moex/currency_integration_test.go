package moex

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestCurrencyReferenceBondsIntegration(t *testing.T) {
	if os.Getenv("MOEX_INTEGRATION") != "1" {
		t.Skip("set MOEX_INTEGRATION=1 to query the live MOEX API")
	}
	client := New(4)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cases := []struct{ code, face string }{
		{"RU000A105KU0", "USD"},
		{"RU000A10A869", "USD"},
		{"RU000A10E8P0", "CNY"},
	}
	codes := make([]string, len(cases))
	for i, tc := range cases {
		codes[i] = tc.code
	}
	results := client.GetBondsInfo(ctx, codes)
	for i, tc := range cases {
		got := results[i]
		if got.Err != nil || got.CleanPrice.Currency != tc.face || got.AccruedInterest.Currency != "RUB" {
			t.Errorf("%s: result %+v", tc.code, got)
		}
	}
	rates, err := client.GetCurrencyRates(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, currency := range []string{"RUB", "USD", "EUR", "CNY"} {
		if rates[currency] <= 0 {
			t.Errorf("%s reference rate unavailable", currency)
		}
	}
	coupons, err := client.GetBondCoupons(ctx, "RU000A10E8P0")
	if err != nil || len(coupons.Coupons) == 0 {
		t.Fatalf("coupon schedule: %+v, %v", coupons, err)
	}
	for _, coupon := range coupons.Coupons {
		if coupon.Amount != nil && coupon.Amount.Currency != "CNY" {
			t.Fatalf("unexpected coupon currency: %+v", coupon)
		}
	}
}
