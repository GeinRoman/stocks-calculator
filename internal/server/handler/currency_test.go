package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/servererrors"
)

type currencyHandlerApp struct {
	App
	err error
}

func (a currencyHandlerApp) GetCurrencyRate(_ context.Context, from, to string) (model.CurrencyRate, error) {
	return model.CurrencyRate{From: from, To: to, Rate: 12}, a.err
}

func TestCurrencyRateEndpoint(t *testing.T) {
	app := &currencyHandlerApp{}
	h := NewHttpServer(app, bondTestTokens{}, 0).Handler
	if w := bondRequest(h, http.MethodGet, "/currencyrate?from=CNY&to=RUB", "", ""); w.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated status %d", w.Code)
	}
	if w := bondRequest(h, http.MethodGet, "/currencyrate?from=CNY", "", "alice"); w.Code != http.StatusBadRequest {
		t.Fatalf("missing currency status %d", w.Code)
	}
	w := bondRequest(h, http.MethodGet, "/currencyrate?from=CNY&to=RUB", "", "alice")
	var got model.CurrencyRate
	if w.Code != http.StatusOK || json.Unmarshal(w.Body.Bytes(), &got) != nil || got.From != "CNY" || got.To != "RUB" || got.Rate != 12 {
		t.Fatalf("status %d, body %s", w.Code, w.Body.String())
	}
	app.err = servererrors.ErrUnsupportedCurrency
	if w := bondRequest(h, http.MethodGet, "/currencyrate?from=XYZ&to=RUB", "", "alice"); w.Code != http.StatusBadRequest {
		t.Fatalf("unsupported currency status %d", w.Code)
	}
}
