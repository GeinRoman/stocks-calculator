package handler

import (
	"context"
	"errors"
	"net/http"
	"stocks_calculator/internal/server/servererrors"
	"strings"
	"time"
)

// @Summary Get bond coupon schedule
// @Description Returns available historical and future coupon amounts in RUB per bond; null means unknown.
// @Tags bonds
// @Produce json
// @Param code query string true "MOEX bond code"
// @Success 200 {object} model.BondCouponsResponse
// @Failure 400 {string} string "Bond code is required"
// @Failure 404 {string} string "Bond not found"
// @Failure 422 {string} string "Unsupported bond currency"
// @Failure 500
// @Failure 504 {string} string "Coupon lookup timed out"
// @Security BearerAuth
// @Router /bondcoupons [get]
func (h *handler) getBondCoupons(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimSpace(r.URL.Query().Get("code"))
	if code == "" {
		http.Error(w, "Bond code is required", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 35*time.Second)
	defer cancel()
	response, err := h.app.GetBondCoupons(ctx, code)
	switch {
	case errors.Is(ctx.Err(), context.DeadlineExceeded), errors.Is(err, context.DeadlineExceeded):
		http.Error(w, "Coupon lookup timed out", http.StatusGatewayTimeout)
	case errors.Is(err, servererrors.ErrMoexBondNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, servererrors.ErrUnsupportedCurrency):
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
	case err != nil:
		w.WriteHeader(http.StatusInternalServerError)
	default:
		writeAndMarshal(w, response)
	}
}
