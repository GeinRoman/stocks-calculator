package handler

import (
	"errors"
	"net/http"
	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/servererrors"
)

// @Summary Get bond groups
// @Tags bond groups
// @Produce json
// @Success 200 {array} model.BondGroup
// @Failure 404 {string} string "No default profile"
// @Failure 500
// @Security BearerAuth
// @Router /getbondgroups [get]
func (h *handler) getBondGroups(w http.ResponseWriter, r *http.Request) {
	userId, httpErr := extractUserId(r)
	if httpErr != nil {
		http.Error(w, httpErr.msg, httpErr.code)
		return
	}
	groups, err := h.app.GetBondGroups(r.Context(), userId)
	switch {
	case errors.Is(err, servererrors.ErrNoDefaultProfile):
		http.Error(w, err.Error(), http.StatusNotFound)
	case err != nil:
		w.WriteHeader(http.StatusInternalServerError)
	default:
		writeAndMarshal(w, groups)
	}
}

// @Summary Add bond groups
// @Tags bond groups
// @Accept json
// @Param body body []model.BondGroup true "Bond groups"
// @Success 200
// @Failure 400 {string} string "Request body is empty or invalid group name"
// @Failure 404 {string} string "No default profile"
// @Failure 409 {string} string "Group already exists"
// @Failure 500
// @Security BearerAuth
// @Router /addbondgroups [post]
func (h *handler) addBondGroups(w http.ResponseWriter, r *http.Request) {
	userId, httpErr := extractUserId(r)
	if httpErr != nil {
		http.Error(w, httpErr.msg, httpErr.code)
		return
	}
	groups, httpErr := readBody[[]model.BondGroup](r)
	if httpErr != nil {
		http.Error(w, httpErr.Error(), httpErr.code)
		return
	}
	if groups == nil || len(*groups) == 0 {
		http.Error(w, "Request body is empty", http.StatusBadRequest)
		return
	}
	err := h.app.AddBondGroups(r.Context(), userId, *groups)
	switch {
	case errors.Is(err, servererrors.ErrInvalidGroupName):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, servererrors.ErrNoDefaultProfile):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, servererrors.ErrGroupExists):
		http.Error(w, err.Error(), http.StatusConflict)
	case err != nil:
		w.WriteHeader(http.StatusInternalServerError)
	default:
		w.WriteHeader(http.StatusOK)
	}
}

// @Summary Remove bond groups
// @Tags bond groups
// @Accept json
// @Param body body []model.BondGroup true "Bond groups"
// @Success 200
// @Failure 400 {string} string "Request body is empty"
// @Failure 404 {string} string "No default profile or group not found"
// @Failure 500
// @Security BearerAuth
// @Router /removebondgroups [delete]
func (h *handler) removeBondGroups(w http.ResponseWriter, r *http.Request) {
	userId, httpErr := extractUserId(r)
	if httpErr != nil {
		http.Error(w, httpErr.msg, httpErr.code)
		return
	}
	groups, httpErr := readBody[[]model.BondGroup](r)
	if httpErr != nil {
		http.Error(w, httpErr.Error(), httpErr.code)
		return
	}
	if groups == nil || len(*groups) == 0 {
		http.Error(w, "Request body is empty", http.StatusBadRequest)
		return
	}
	err := h.app.RemoveBondGroups(r.Context(), userId, *groups)
	switch {
	case errors.Is(err, servererrors.ErrNoDefaultProfile), errors.Is(err, servererrors.ErrGroupNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case err != nil:
		w.WriteHeader(http.StatusInternalServerError)
	default:
		w.WriteHeader(http.StatusOK)
	}
}

// @Summary Rename bond group
// @Tags bond groups
// @Accept json
// @Param body body model.RenameBondGroupBody true "Rename bond group body"
// @Success 200
// @Failure 400 {string} string "Malformed request body"
// @Failure 404 {string} string "No default profile or group not found"
// @Failure 500
// @Security BearerAuth
// @Router /renamebondgroup [patch]
func (h *handler) renameBondGroup(w http.ResponseWriter, r *http.Request) {
	userId, httpErr := extractUserId(r)
	if httpErr != nil {
		http.Error(w, httpErr.msg, httpErr.code)
		return
	}
	group, httpErr := readBody[model.RenameBondGroupBody](r)
	if httpErr != nil {
		http.Error(w, httpErr.Error(), httpErr.code)
		return
	}
	if group.NameNew == "" || group.NameOld == "" {
		http.Error(w, "Malformed request body", http.StatusBadRequest)
		return
	}
	err := h.app.RenameBondGroup(r.Context(), userId, *group)
	switch {
	case errors.Is(err, servererrors.ErrNoDefaultProfile), errors.Is(err, servererrors.ErrGroupNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case err != nil:
		w.WriteHeader(http.StatusInternalServerError)
	default:
		w.WriteHeader(http.StatusOK)
	}
}

// @Summary Update bond weights
// @Tags bond groups
// @Accept json
// @Param body body []model.BondGroup true "Bond groups"
// @Success 200
// @Failure 400 {string} string "Request body is empty or invalid weights"
// @Failure 404 {string} string "No default profile or group not found"
// @Failure 500
// @Security BearerAuth
// @Router /updatebondweights [patch]
func (h *handler) updateBondWeights(w http.ResponseWriter, r *http.Request) {
	userId, httpErr := extractUserId(r)
	if httpErr != nil {
		http.Error(w, httpErr.msg, httpErr.code)
		return
	}
	groups, httpErr := readBody[[]model.BondGroup](r)
	if httpErr != nil {
		http.Error(w, httpErr.Error(), httpErr.code)
		return
	}
	if groups == nil || len(*groups) == 0 {
		http.Error(w, "Request body is empty", http.StatusBadRequest)
		return
	}
	err := h.app.UpdateBondWeights(r.Context(), userId, *groups)
	switch {
	case errors.Is(err, servererrors.ErrInvalidWeights):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, servererrors.ErrNoDefaultProfile), errors.Is(err, servererrors.ErrGroupNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case err != nil:
		w.WriteHeader(http.StatusInternalServerError)
	default:
		w.WriteHeader(http.StatusOK)
	}
}
