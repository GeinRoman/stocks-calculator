package handler

import (
	"errors"
	"net/http"
	"stocks_calculator/internal/server/servererrors"
	"strings"
)

// @Summary Get currency conversion reference rate
// @Description One unit of from equals rate units of to. Supports RUB, USD, EUR and CNY.
// @Description USD and EUR use CBR reference values published by MOEX; CNY uses the previous MOEX weighted-average rate.
// @Tags currency
// @Produce json
// @Param from query string true "Source currency"
// @Param to query string true "Target currency"
// @Success 200 {object} model.CurrencyRate
// @Failure 400 {string} string "Missing or unsupported currency"
// @Failure 500
// @Security BearerAuth
// @Router /currencyrate [get]
func (h *handler) getCurrencyRate(w http.ResponseWriter, r *http.Request) {
	from, to := strings.TrimSpace(r.URL.Query().Get("from")), strings.TrimSpace(r.URL.Query().Get("to"))
	if from == "" || to == "" {
		http.Error(w, "Both currencies are required", http.StatusBadRequest)
		return
	}
	result, err := h.app.GetCurrencyRate(r.Context(), from, to)
	switch {
	case errors.Is(err, servererrors.ErrUnsupportedCurrency):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case err != nil:
		w.WriteHeader(http.StatusInternalServerError)
	default:
		writeAndMarshal(w, result)
	}
}
