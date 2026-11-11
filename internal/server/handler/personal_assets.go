package handler

import (
	"database/sql"
	"errors"
	"net/http"

	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/servererrors"
)

type personalAssetsResponse struct {
	Deposits    []model.Deposit    `json:"deposits"`
	CreditCards []model.CreditCard `json:"credit_cards"`
	Insurances  []model.Insurance  `json:"insurances"`
}

type personalAssetID struct {
	ID int64 `json:"id"`
}

func readPersonalAssetBody[T any](w http.ResponseWriter, r *http.Request) (int, *T, bool) {
	id, e := extractUserId(r)
	if e != nil {
		http.Error(w, e.msg, e.code)
		return 0, nil, false
	}
	body, e := readBody[T](r)
	if e != nil {
		http.Error(w, e.msg, e.code)
		return 0, nil, false
	}
	return id, body, true
}

// @Summary List personal assets
// @Tags personal-assets
// @Produce json
// @Success 200 {object} personalAssetsResponse
// @Failure 404 {string} string
// @Failure 500
// @Security BearerAuth
// @Router /getpersonalassets [get]
func (h *handler) getPersonalAssets(w http.ResponseWriter, r *http.Request) {
	id, e := extractUserId(r)
	if e != nil {
		http.Error(w, e.msg, e.code)
		return
	}
	ctx, err := h.app.PersonalAssetProfileContext(r.Context(), id)
	if err != nil {
		personalAssetError(w, err)
		return
	}
	deposits, err := h.app.GetDeposits(ctx, id)
	if err != nil {
		personalAssetError(w, err)
		return
	}
	cards, err := h.app.GetCreditCards(ctx, id)
	if err != nil {
		personalAssetError(w, err)
		return
	}
	insurances, err := h.app.GetInsurances(ctx, id)
	if err != nil {
		personalAssetError(w, err)
		return
	}
	if deposits == nil {
		deposits = []model.Deposit{}
	}
	if cards == nil {
		cards = []model.CreditCard{}
	}
	if insurances == nil {
		insurances = []model.Insurance{}
	}
	w.Header().Set("Content-Type", "application/json")
	writeAndMarshal(w, personalAssetsResponse{deposits, cards, insurances})
}

func personalAssetError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, servererrors.ErrNoDefaultProfile), errors.Is(err, model.ErrCreditCardNotFound), errors.Is(err, sql.ErrNoRows):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, model.ErrInvalidCreditCard), errors.Is(err, servererrors.ErrInvalidContract):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}
