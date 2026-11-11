package handler

import (
	"net/http"

	"stocks_calculator/internal/model"
)

// @Summary Add a credit card
// @Description Date fields are strings.
// @Tags personal-assets
// @Accept json
// @Param card body model.CreditCard true "Credit card"
// @Success 200
// @Failure 400 {string} string
// @Failure 404 {string} string
// @Failure 500
// @Security BearerAuth
// @Router /addcreditcard [post]
func (h *handler) addCreditCard(w http.ResponseWriter, r *http.Request) {
	id, body, ok := readPersonalAssetBody[model.CreditCard](w, r)
	if !ok {
		return
	}
	if err := h.app.AddCreditCard(r.Context(), id, *body); err != nil {
		personalAssetError(w, err)
	}
}

// @Summary Update a credit card
// @Description Date fields are strings.
// @Tags personal-assets
// @Accept json
// @Param card body model.CreditCardUpdate true "Credit card update"
// @Success 200
// @Failure 400 {string} string
// @Failure 404 {string} string
// @Failure 500
// @Security BearerAuth
// @Router /updatecreditcard [patch]
func (h *handler) updateCreditCard(w http.ResponseWriter, r *http.Request) {
	id, body, ok := readPersonalAssetBody[model.CreditCardUpdate](w, r)
	if !ok {
		return
	}
	if err := h.app.UpdateCreditCard(r.Context(), id, *body); err != nil {
		personalAssetError(w, err)
	}
}

// @Summary Remove a credit card
// @Tags personal-assets
// @Accept json
// @Param card body personalAssetID true "Credit card ID"
// @Success 200
// @Failure 400 {string} string
// @Failure 404 {string} string
// @Failure 500
// @Security BearerAuth
// @Router /removecreditcard [delete]
func (h *handler) removeCreditCard(w http.ResponseWriter, r *http.Request) {
	id, body, ok := readPersonalAssetBody[personalAssetID](w, r)
	if !ok {
		return
	}
	if err := h.app.RemoveCreditCard(r.Context(), id, body.ID); err != nil {
		personalAssetError(w, err)
	}
}
