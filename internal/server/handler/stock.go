package handler

import (
	"errors"
	"net/http"
	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/app"
)

// @Summary		Add stocks
// @Tags		stocks
// @Accept		json
// @Param		body	body	[]model.Stock	true	"Stocks"
// @Success	200
// @Failure	400	{string}	string	"Request body is empty / wrong stock group / wrong stock amount"
// @Failure	404	{string}	string	"Group not found"
// @Failure	500
// @Security	BearerAuth
// @Router		/addstocks [post]
func (h *handler) addStocks(w http.ResponseWriter, r *http.Request) {
	userId, httpErr := extractUserId(r)
	if httpErr != nil {
		http.Error(w, httpErr.msg, httpErr.code)
		return
	}

	stocks, httpErr := readBody[[]model.Stock](r)
	if httpErr != nil {
		http.Error(w, httpErr.Error(), httpErr.code)
		return
	}

	if stocks == nil || len(*stocks) == 0 {
		http.Error(w, "Request body is empty", http.StatusBadRequest)
		return
	}

	err := h.app.AddStocks(r.Context(), userId, *stocks)
	switch {
	case errors.Is(err, app.ErrGroupNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	case errors.Is(err, app.ErrWrongStockGroup):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case errors.Is(err, app.ErrWrongStockAmount):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case err != nil:
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// @Summary		Remove stocks
// @Tags		stocks
// @Accept		json
// @Param		body	body	[]model.Stock	true	"Stocks"
// @Success	200
// @Failure	400	{string}	string	"Request body is empty"
// @Failure	404	{string}	string	"Stock not found"
// @Failure	500
// @Security	BearerAuth
// @Router		/removestocks [delete]
func (h *handler) removeStocks(w http.ResponseWriter, r *http.Request) {
	userId, httpErr := extractUserId(r)
	if httpErr != nil {
		http.Error(w, httpErr.msg, httpErr.code)
		return
	}

	stocks, httpErr := readBody[[]model.Stock](r)
	if httpErr != nil {
		http.Error(w, httpErr.Error(), httpErr.code)
		return
	}

	if stocks == nil || len(*stocks) == 0 {
		http.Error(w, "Request body is empty", http.StatusBadRequest)
		return
	}

	err := h.app.RemoveStocks(r.Context(), userId, *stocks)
	switch {
	case errors.Is(err, app.ErrStockNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	case errors.Is(err, app.ErrWrongStockGroup):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case err != nil:
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// @Summary		Get stocks
// @Tags		stocks
// @Produce	json
// @Success	200	{array}		model.Stock
// @Failure	500
// @Security	BearerAuth
// @Router		/getstocks [get]
func (h *handler) getStocks(w http.ResponseWriter, r *http.Request) {
	userId, httpErr := extractUserId(r)
	if httpErr != nil {
		http.Error(w, httpErr.msg, httpErr.code)
		return
	}

	stocks, err := h.app.GetStocks(r.Context(), userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	writeAndMarshal(w, stocks)
}
