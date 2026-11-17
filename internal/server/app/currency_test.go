package app

import (
	"context"
	"errors"
	"testing"

	"stocks_calculator/internal/server/servererrors"
)

type currencyTestMoex struct {
	MoexApi
	rates map[string]float64
	err   error
	calls int
}

func (m *currencyTestMoex) GetCurrencyRates(context.Context) (map[string]float64, error) {
	m.calls++
	return m.rates, m.err
}

func TestGetCurrencyRate(t *testing.T) {
	m := &currencyTestMoex{rates: map[string]float64{"RUB": 1, "USD": 80, "EUR": 100, "CNY": 10}}
	a := New(nil, nil, m)
	for _, tc := range []struct {
		from, to string
		want     float64
	}{{"usd", "RUB", 80}, {"RUB", "CNY", .1}, {"USD", "EUR", .8}, {"CNY", "CNY", 1}} {
		result, err := a.GetCurrencyRate(context.Background(), tc.from, tc.to)
		if err != nil || result.Rate != tc.want {
			t.Fatalf("%s -> %s: %+v, %v", tc.from, tc.to, result, err)
		}
	}
	if m.calls != 3 {
		t.Fatalf("MOEX calls = %d, want 3", m.calls)
	}
	if _, err := a.GetCurrencyRate(context.Background(), "CAD", "CAD"); !errors.Is(err, servererrors.ErrUnsupportedCurrency) {
		t.Fatalf("unsupported currency: %v", err)
	}
	delete(m.rates, "CNY")
	if _, err := a.GetCurrencyRate(context.Background(), "CNY", "RUB"); !errors.Is(err, servererrors.ErrMoexUnhandled) {
		t.Fatalf("missing rate: %v", err)
	}
}
