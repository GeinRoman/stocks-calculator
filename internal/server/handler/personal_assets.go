package handler

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/servererrors"
)

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
// @Success 200 {object} model.PersonalAssets
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
	assets, err := h.personalAssets(r.Context(), id)
	if err != nil {
		personalAssetError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	writeAndMarshal(w, assets)
}

func (h *handler) personalAssets(ctx context.Context, id int) (model.PersonalAssets, error) {
	ctx, err := h.app.PersonalAssetProfileContext(ctx, id)
	if err != nil {
		return model.PersonalAssets{}, err
	}
	deposits, err := h.app.GetDeposits(ctx, id)
	if err != nil {
		return model.PersonalAssets{}, err
	}
	cards, err := h.app.GetCreditCards(ctx, id)
	if err != nil {
		return model.PersonalAssets{}, err
	}
	insurances, err := h.app.GetInsurances(ctx, id)
	if err != nil {
		return model.PersonalAssets{}, err
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
	return model.PersonalAssets{Deposits: deposits, CreditCards: cards, Insurances: insurances}, nil
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
