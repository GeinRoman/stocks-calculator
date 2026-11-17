package moex

import (
	"context"
	"fmt"
	"math"
)

// GetCurrencyRates returns MOEX reference rates in RUB per currency unit.
func (m *moex) GetCurrencyRates(ctx context.Context) (map[string]float64, error) {
	var response struct {
		CBRF bondTable `json:"cbrf"`
		WAP  bondTable `json:"wap_rates"`
	}
	if err := m.bondRequest(ctx, "/statistics/engines/currency/markets/selt/rates.json", nil, &response); err != nil {
		return nil, err
	}
	rates := map[string]float64{"RUB": 1}
	cbrf, err := response.CBRF.rows("CBRF_USD_LAST", "CBRF_EUR_LAST")
	if err != nil {
		return nil, err
	}
	if len(cbrf) > 1 {
		return nil, fmt.Errorf("invalid MOEX reference rates")
	}
	if len(cbrf) == 1 {
		for _, pair := range []struct{ currency, column string }{{"USD", "cbrf_usd_last"}, {"EUR", "cbrf_eur_last"}} {
			if price, ok := cbrf[0][pair.column].(float64); ok && price > 0 && !math.IsInf(price, 0) {
				rates[pair.currency] = price
			}
		}
	}
	wap, err := response.WAP.rows("secid", "price", "nominal")
	if err != nil {
		return nil, err
	}
	for _, row := range wap {
		if row["secid"] != "CNYRUB_TOM" {
			continue
		}
		price, pok := row["price"].(float64)
		nominal, nok := row["nominal"].(float64)
		if pok && nok && price > 0 && nominal > 0 && !math.IsInf(price, 0) && !math.IsInf(nominal, 0) {
			rates["CNY"] = price / nominal
		}
	}
	return rates, nil
}
