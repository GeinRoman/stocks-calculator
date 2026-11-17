package handler

import (
	"errors"
	"net/http"
	"stocks_calculator/internal/server/servererrors"
	"strconv"
)

// @Summary Rebalance bond portfolio
// @Tags rebalance
// @Produce json
// @Param nosell query bool true "If true, rebalance by buying only without selling"
// @Param valdiff query number true "Desired value difference; must be positive if nosell is true"
// @Success 200 {object} model.BondRebalanceResponse
// @Failure 400 {string} string "Missing or invalid params"
// @Failure 404 {string} string "No default profile"
// @Failure 428 {string} string "Bonds or groups not found, or groups not weighted"
// @Failure 500
// @Security BearerAuth
// @Router /rebalancebonds [get]
func (h *handler) rebalanceBonds(w http.ResponseWriter, r *http.Request) {
	userId, httpErr := extractUserId(r)
	if httpErr != nil {
		http.Error(w, httpErr.msg, httpErr.code)
		return
	}
	nosell, err := strconv.ParseBool(r.URL.Query().Get("nosell"))
	if err != nil {
		http.Error(w, "Missing nosell param", http.StatusBadRequest)
		return
	}
	valueDiff, err := strconv.ParseFloat(r.URL.Query().Get("valdiff"), 64)
	if err != nil {
		http.Error(w, "Missing valdiff param", http.StatusBadRequest)
		return
	}
	if nosell && valueDiff <= 0 {
		http.Error(w, "Invalid params values", http.StatusBadRequest)
		return
	}
	response, err := h.app.RebalanceBonds(r.Context(), userId, nosell, valueDiff)
	switch {
	case errors.Is(err, servererrors.ErrNoDefaultProfile):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, servererrors.ErrBondsNotFound), errors.Is(err, servererrors.ErrGroupsNotFound), errors.Is(err, servererrors.ErrGroupsAreNotWeighted):
		http.Error(w, err.Error(), http.StatusPreconditionRequired)
	case err != nil:
		w.WriteHeader(http.StatusInternalServerError)
	default:
		writeAndMarshal(w, response)
	}
}
