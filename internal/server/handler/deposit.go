package handler

import (
	"net/http"

	"stocks_calculator/internal/model"
)

// @Summary Add a deposit
// @Description Date fields are strings.
// @Tags personal-assets
// @Accept json
// @Param deposit body model.Deposit true "Deposit"
// @Success 200
// @Failure 400 {string} string
// @Failure 404 {string} string
// @Failure 500
// @Security BearerAuth
// @Router /adddeposit [post]
func (h *handler) addDeposit(w http.ResponseWriter, r *http.Request) {
	id, body, ok := readPersonalAssetBody[model.Deposit](w, r)
	if !ok {
		return
	}
	if err := h.app.AddDeposit(r.Context(), id, *body); err != nil {
		personalAssetError(w, err)
	}
}

// @Summary Update a deposit
// @Description Date fields are strings.
// @Tags personal-assets
// @Accept json
// @Param deposit body model.DepositUpdate true "Deposit update"
// @Success 200
// @Failure 400 {string} string
// @Failure 404 {string} string
// @Failure 500
// @Security BearerAuth
// @Router /updatedeposit [patch]
func (h *handler) updateDeposit(w http.ResponseWriter, r *http.Request) {
	id, body, ok := readPersonalAssetBody[model.DepositUpdate](w, r)
	if !ok {
		return
	}
	if err := h.app.UpdateDeposit(r.Context(), id, *body); err != nil {
		personalAssetError(w, err)
	}
}

// @Summary Remove a deposit
// @Tags personal-assets
// @Accept json
// @Param deposit body personalAssetID true "Deposit ID"
// @Success 200
// @Failure 400 {string} string
// @Failure 404 {string} string
// @Failure 500
// @Security BearerAuth
// @Router /removedeposit [delete]
func (h *handler) removeDeposit(w http.ResponseWriter, r *http.Request) {
	id, body, ok := readPersonalAssetBody[personalAssetID](w, r)
	if !ok {
		return
	}
	if err := h.app.RemoveDeposit(r.Context(), id, body.ID); err != nil {
		personalAssetError(w, err)
	}
}

// @Summary Close a deposit
// @Description Date fields are strings.
// @Tags personal-assets
// @Accept json
// @Param deposit body model.DepositClose true "Deposit close"
// @Success 200
// @Failure 400 {string} string
// @Failure 404 {string} string
// @Failure 500
// @Security BearerAuth
// @Router /closedeposit [post]
func (h *handler) closeDeposit(w http.ResponseWriter, r *http.Request) {
	id, body, ok := readPersonalAssetBody[model.DepositClose](w, r)
	if !ok {
		return
	}
	if err := h.app.CloseDeposit(r.Context(), id, *body); err != nil {
		personalAssetError(w, err)
	}
}
