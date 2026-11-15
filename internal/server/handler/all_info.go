package handler

import (
	"errors"
	"net/http"

	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/servererrors"
)

// @Summary Get all portfolio information
// @Description Returns all profiles and the default profile's assets and groups with weights, including current MOEX data for stocks and bonds.
// @Tags portfolio
// @Produce json
// @Success 200 {object} model.AllInfo
// @Failure 401 {string} string
// @Failure 404 {string} string "No default profile"
// @Failure 500
// @Security BearerAuth
// @Router /getall [get]
func (h *handler) getAll(w http.ResponseWriter, r *http.Request) {
	id, httpErr := extractUserId(r)
	if httpErr != nil {
		http.Error(w, httpErr.msg, httpErr.code)
		return
	}
	ctx := r.Context()
	var result model.AllInfo
	var err error
	if result.PersonalAssets, err = h.personalAssets(ctx, id); err != nil {
		allInfoError(w, err)
		return
	}
	if result.Bonds, err = h.app.GetBonds(ctx, id); err != nil {
		allInfoError(w, err)
		return
	}
	if result.BondGroups, err = h.app.GetBondGroups(ctx, id); err != nil {
		allInfoError(w, err)
		return
	}
	if result.CryptoAssets, err = h.app.GetCryptoAssets(ctx, id); err != nil {
		allInfoError(w, err)
		return
	}
	if result.CryptoGroups, err = h.app.GetCryptoGroups(ctx, id); err != nil {
		allInfoError(w, err)
		return
	}
	if result.Stocks, err = h.app.GetStocks(ctx, id); err != nil {
		allInfoError(w, err)
		return
	}
	if result.StockGroups, err = h.app.GetGroups(ctx, id); err != nil {
		allInfoError(w, err)
		return
	}
	if result.Profiles, err = h.app.GetProfiles(ctx, id); err != nil {
		allInfoError(w, err)
		return
	}
	if result.Bonds == nil {
		result.Bonds = []model.Bond{}
	}
	if result.BondGroups == nil {
		result.BondGroups = []model.BondGroup{}
	}
	if result.CryptoAssets == nil {
		result.CryptoAssets = []model.CryptoAsset{}
	}
	if result.CryptoGroups == nil {
		result.CryptoGroups = []model.CryptoGroup{}
	}
	if result.Stocks == nil {
		result.Stocks = []model.Stock{}
	}
	if result.StockGroups == nil {
		result.StockGroups = []model.StockGroup{}
	}
	if result.Profiles == nil {
		result.Profiles = []model.Profile{}
	}
	w.Header().Set("Content-Type", "application/json")
	writeAndMarshal(w, result)
}

func allInfoError(w http.ResponseWriter, err error) {
	if errors.Is(err, servererrors.ErrNoDefaultProfile) {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusInternalServerError)
}
