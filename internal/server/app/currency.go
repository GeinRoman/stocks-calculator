package app

import (
	"context"
	"fmt"
	"math"
	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/servererrors"
	"strings"
)

func supportedCurrency(currency string) bool {
	switch currency {
	case "RUB", "USD", "EUR", "CNY":
		return true
	default:
		return false
	}
}

func (a *app) GetCurrencyRate(ctx context.Context, from, to string) (model.CurrencyRate, error) {
	from, to = strings.ToUpper(strings.TrimSpace(from)), strings.ToUpper(strings.TrimSpace(to))
	if !supportedCurrency(from) || !supportedCurrency(to) {
		return model.CurrencyRate{}, servererrors.ErrUnsupportedCurrency
	}
	result := model.CurrencyRate{From: from, To: to, Rate: 1}
	if from == to {
		return result, nil
	}
	rates, err := a.moex.GetCurrencyRates(ctx)
	if err != nil {
		return model.CurrencyRate{}, fmt.Errorf("%w: %w", servererrors.ErrMoexUnhandled, err)
	}
	fromRate, fromOK := rates[from]
	toRate, toOK := rates[to]
	if !fromOK || !toOK || fromRate <= 0 || toRate <= 0 || math.IsNaN(fromRate) || math.IsNaN(toRate) || math.IsInf(fromRate, 0) || math.IsInf(toRate, 0) {
		return model.CurrencyRate{}, fmt.Errorf("%w: currency reference rate unavailable", servererrors.ErrMoexUnhandled)
	}
	result.Rate = fromRate / toRate
	if result.Rate <= 0 || math.IsInf(result.Rate, 0) || math.IsNaN(result.Rate) {
		return model.CurrencyRate{}, fmt.Errorf("%w: invalid conversion rate", servererrors.ErrMoexUnhandled)
	}
	return result, nil
}
