package handler

import (
	"errors"
	"net/http"
	"stocks_calculator/internal/server/servererrors"
	"strconv"
)

// @Summary		Rebalance portfolio
// @Tags		rebalance
// @Produce	json
// @Param		nosell	query	bool	true	"If true, rebalance by buying only without selling"
// @Param		valdiff	query	number	true	"Value difference to rebalance with (required if nosell is true, must be > 0)"
// @Success	200		{object}	model.RebalanceResponse
// @Failure	400		{string}	string	"Missing or invalid params"
// @Failure	428		{string}	string	"Stocks/groups not found or groups not weighted"
// @Failure	500
// @Security	BearerAuth
// @Router		/rebalance [get]
func (h *handler) rebalance(w http.ResponseWriter, r *http.Request) {
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

	response, err := h.app.Rebalance(r.Context(), userId, nosell, valueDiff)
	switch {
	case errors.Is(err, servererrors.ErrStocksNotFound):
		http.Error(w, err.Error(), http.StatusPreconditionRequired)
		return
	case errors.Is(err, servererrors.ErrGroupsNotFound):
		http.Error(w, err.Error(), http.StatusPreconditionRequired)
		return
	case errors.Is(err, servererrors.ErrGroupsAreNotWeighted):
		http.Error(w, err.Error(), http.StatusPreconditionRequired)
		return
	case err != nil:
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	writeAndMarshal(w, response)
}
