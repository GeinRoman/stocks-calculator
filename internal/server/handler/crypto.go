package handler

import (
	"errors"
	"net/http"

	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/servererrors"
)

// @Summary List crypto groups
// @Tags crypto
// @Produce json
// @Success 200 {array} model.CryptoGroup
// @Failure 404 {string} string
// @Failure 500
// @Security BearerAuth
// @Router /getcryptogroups [get]
func (h *handler) getCryptoGroups(w http.ResponseWriter, r *http.Request) {
	id, e := extractUserId(r)
	if e != nil {
		http.Error(w, e.msg, e.code)
		return
	}
	v, err := h.app.GetCryptoGroups(r.Context(), id)
	if err != nil {
		cryptoError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	writeAndMarshal(w, v)
}

// @Summary List crypto assets
// @Tags crypto
// @Produce json
// @Success 200 {array} model.CryptoAsset
// @Failure 404 {string} string
// @Failure 500
// @Security BearerAuth
// @Router /getcryptoassets [get]
func (h *handler) getCryptoAssets(w http.ResponseWriter, r *http.Request) {
	id, e := extractUserId(r)
	if e != nil {
		http.Error(w, e.msg, e.code)
		return
	}
	v, err := h.app.GetCryptoAssets(r.Context(), id)
	if err != nil {
		cryptoError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	writeAndMarshal(w, v)
}

// @Summary Add crypto groups
// @Description New groups are created with weight zero, regardless of the supplied weight. Update weights separately.
// @Tags crypto
// @Accept json
// @Param groups body []model.CryptoGroup true "Groups"
// @Success 200
// @Failure 400 {string} string
// @Failure 404 {string} string
// @Failure 409 {string} string
// @Failure 500
// @Security BearerAuth
// @Router /addcryptogroups [post]
func (h *handler) addCryptoGroups(w http.ResponseWriter, r *http.Request) {
	id, groups, ok := readCryptoBody[[]model.CryptoGroup](w, r)
	if !ok {
		return
	}
	if len(*groups) == 0 {
		cryptoError(w, servererrors.ErrInvalidCrypto)
		return
	}
	if err := h.app.AddCryptoGroups(r.Context(), id, *groups); err != nil {
		cryptoError(w, err)
	}
}

// @Summary Remove crypto groups
// @Tags crypto
// @Accept json
// @Param groups body []model.CryptoGroup true "Groups"
// @Success 200
// @Failure 400 {string} string
// @Failure 404 {string} string
// @Failure 409 {string} string
// @Failure 500
// @Security BearerAuth
// @Router /removecryptogroups [delete]
func (h *handler) removeCryptoGroups(w http.ResponseWriter, r *http.Request) {
	id, groups, ok := readCryptoBody[[]model.CryptoGroup](w, r)
	if !ok {
		return
	}
	if len(*groups) == 0 {
		cryptoError(w, servererrors.ErrInvalidCrypto)
		return
	}
	if err := h.app.RemoveCryptoGroups(r.Context(), id, *groups); err != nil {
		cryptoError(w, err)
	}
}

// @Summary Rename a crypto group
// @Tags crypto
// @Accept json
// @Param group body model.RenameCryptoGroupBody true "Group names"
// @Success 200
// @Failure 400 {string} string
// @Failure 404 {string} string
// @Failure 409 {string} string
// @Failure 500
// @Security BearerAuth
// @Router /renamecryptogroup [patch]
func (h *handler) renameCryptoGroup(w http.ResponseWriter, r *http.Request) {
	id, group, ok := readCryptoBody[model.RenameCryptoGroupBody](w, r)
	if !ok {
		return
	}
	if err := h.app.RenameCryptoGroup(r.Context(), id, *group); err != nil {
		cryptoError(w, err)
	}
}

// @Summary Update a crypto asset
// @Tags crypto
// @Accept json
// @Param asset body model.CryptoAssetUpdate true "Asset"
// @Success 200
// @Failure 400 {string} string
// @Failure 404 {string} string
// @Failure 409 {string} string
// @Failure 500
// @Security BearerAuth
// @Router /updatecryptoasset [patch]
func (h *handler) updateCryptoAsset(w http.ResponseWriter, r *http.Request) {
	id, asset, ok := readCryptoBody[model.CryptoAssetUpdate](w, r)
	if !ok {
		return
	}
	if err := h.app.UpdateCryptoAsset(r.Context(), id, *asset); err != nil {
		cryptoError(w, err)
	}
}

// @Summary Update crypto group weights
// @Tags crypto
// @Accept json
// @Param groups body []model.CryptoGroup true "Groups"
// @Success 200
// @Failure 400 {string} string
// @Failure 404 {string} string
// @Failure 409 {string} string
// @Failure 500
// @Security BearerAuth
// @Router /updatecryptoweights [patch]
func (h *handler) updateCryptoWeights(w http.ResponseWriter, r *http.Request) {
	id, groups, ok := readCryptoBody[[]model.CryptoGroup](w, r)
	if !ok {
		return
	}
	if len(*groups) == 0 {
		cryptoError(w, servererrors.ErrInvalidCrypto)
		return
	}
	if err := h.app.UpdateCryptoWeights(r.Context(), id, *groups); err != nil {
		cryptoError(w, err)
	}
}

// @Summary Buy crypto
// @Tags crypto
// @Accept json
// @Param operations body []model.CryptoOperation true "Operations"
// @Success 200
// @Failure 400 {string} string
// @Failure 404 {string} string
// @Failure 409 {string} string
// @Failure 500
// @Security BearerAuth
// @Router /buycrypto [post]
func (h *handler) buyCrypto(w http.ResponseWriter, r *http.Request) {
	id, operations, ok := readCryptoBody[[]model.CryptoOperation](w, r)
	if !ok {
		return
	}
	if len(*operations) == 0 {
		cryptoError(w, servererrors.ErrInvalidCrypto)
		return
	}
	if err := h.app.BuyCrypto(r.Context(), id, *operations); err != nil {
		cryptoError(w, err)
	}
}

// @Summary Sell crypto
// @Tags crypto
// @Accept json
// @Param operations body []model.CryptoOperation true "Operations"
// @Success 200
// @Failure 400 {string} string
// @Failure 404 {string} string
// @Failure 409 {string} string
// @Failure 500
// @Security BearerAuth
// @Router /sellcrypto [delete]
func (h *handler) sellCrypto(w http.ResponseWriter, r *http.Request) {
	id, operations, ok := readCryptoBody[[]model.CryptoOperation](w, r)
	if !ok {
		return
	}
	if len(*operations) == 0 {
		cryptoError(w, servererrors.ErrInvalidCrypto)
		return
	}
	if err := h.app.SellCrypto(r.Context(), id, *operations); err != nil {
		cryptoError(w, err)
	}
}

func readCryptoBody[T any](w http.ResponseWriter, r *http.Request) (int, *T, bool) {
	id, e := extractUserId(r)
	if e != nil {
		http.Error(w, e.msg, e.code)
		return 0, nil, false
	}
	v, e := readBody[T](r)
	if e != nil {
		http.Error(w, e.Error(), e.code)
		return 0, nil, false
	}
	if v == nil {
		http.Error(w, "Request body is empty", http.StatusBadRequest)
		return 0, nil, false
	}
	return id, v, true
}

func cryptoError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, servererrors.ErrNoDefaultProfile), errors.Is(err, servererrors.ErrGroupNotFound), errors.Is(err, servererrors.ErrCryptoNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, servererrors.ErrGroupExists), errors.Is(err, servererrors.ErrCryptoExists), errors.Is(err, servererrors.ErrCryptoGroupNotEmpty):
		http.Error(w, err.Error(), http.StatusConflict)
	case errors.Is(err, servererrors.ErrInvalidCrypto), errors.Is(err, servererrors.ErrInvalidGroupName), errors.Is(err, servererrors.ErrInvalidWeights):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}
