package handler

import (
	"net/http"

	"stocks_calculator/internal/model"
)

// @Summary Add insurance
// @Description Date fields are strings.
// @Tags personal-assets
// @Accept json
// @Param insurance body model.Insurance true "Insurance"
// @Success 200
// @Failure 400 {string} string
// @Failure 404 {string} string
// @Failure 500
// @Security BearerAuth
// @Router /addinsurance [post]
func (h *handler) addInsurance(w http.ResponseWriter, r *http.Request) {
	id, body, ok := readPersonalAssetBody[model.Insurance](w, r)
	if !ok {
		return
	}
	if err := h.app.AddInsurance(r.Context(), id, *body); err != nil {
		personalAssetError(w, err)
	}
}

// @Summary Update insurance
// @Description Date fields are strings.
// @Tags personal-assets
// @Accept json
// @Param insurance body model.InsuranceUpdate true "Insurance update"
// @Success 200
// @Failure 400 {string} string
// @Failure 404 {string} string
// @Failure 500
// @Security BearerAuth
// @Router /updateinsurance [patch]
func (h *handler) updateInsurance(w http.ResponseWriter, r *http.Request) {
	id, body, ok := readPersonalAssetBody[model.InsuranceUpdate](w, r)
	if !ok {
		return
	}
	if err := h.app.UpdateInsurance(r.Context(), id, *body); err != nil {
		personalAssetError(w, err)
	}
}

// @Summary Remove insurance
// @Tags personal-assets
// @Accept json
// @Param insurance body personalAssetID true "Insurance ID"
// @Success 200
// @Failure 400 {string} string
// @Failure 404 {string} string
// @Failure 500
// @Security BearerAuth
// @Router /removeinsurance [delete]
func (h *handler) removeInsurance(w http.ResponseWriter, r *http.Request) {
	id, body, ok := readPersonalAssetBody[personalAssetID](w, r)
	if !ok {
		return
	}
	if err := h.app.RemoveInsurance(r.Context(), id, body.ID); err != nil {
		personalAssetError(w, err)
	}
}

// @Summary Close insurance
// @Description Date fields are strings.
// @Tags personal-assets
// @Accept json
// @Param insurance body model.InsuranceClose true "Insurance close"
// @Success 200
// @Failure 400 {string} string
// @Failure 404 {string} string
// @Failure 500
// @Security BearerAuth
// @Router /closeinsurance [post]
func (h *handler) closeInsurance(w http.ResponseWriter, r *http.Request) {
	id, body, ok := readPersonalAssetBody[model.InsuranceClose](w, r)
	if !ok {
		return
	}
	if err := h.app.CloseInsurance(r.Context(), id, *body); err != nil {
		personalAssetError(w, err)
	}
}
