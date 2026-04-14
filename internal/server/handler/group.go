package handler

import (
	"errors"
	"net/http"
	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/app"
)

// @Summary		Get groups
// @Tags		groups
// @Produce	json
// @Success	200	{array}		model.Group
// @Failure	404	{string}	string	"No default profile"
// @Failure	500
// @Security	BearerAuth
// @Router		/getgroups [get]
func (h *handler) getGroups(w http.ResponseWriter, r *http.Request) {
	userId, httpErr := extractUserId(r)
	if httpErr != nil {
		http.Error(w, httpErr.msg, httpErr.code)
		return
	}

	groups, err := h.app.GetGroups(r.Context(), userId)
	switch {
	case errors.Is(err, app.ErrNoDefaultProfile):
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	case err != nil:
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	writeAndMarshal(w, groups)
}

// @Summary		Add groups
// @Tags		groups
// @Accept		json
// @Param		body	body	[]model.Group	true	"Groups"
// @Success	200
// @Failure	400	{string}	string	"Request body is empty"
// @Failure	404	{string}	string	"No default profile"
// @Failure	409	{string}	string	"Group already exists"
// @Failure	500
// @Security	BearerAuth
// @Router		/addgroups [post]
func (h *handler) addGroups(w http.ResponseWriter, r *http.Request) {
	userId, httpErr := extractUserId(r)
	if httpErr != nil {
		http.Error(w, httpErr.msg, httpErr.code)
		return
	}

	groups, httpErr := readBody[[]model.Group](r)
	if httpErr != nil {
		http.Error(w, httpErr.Error(), httpErr.code)
		return
	}
	if groups == nil || len(*groups) == 0 {
		http.Error(w, "Request body is empty", http.StatusBadRequest)
		return
	}

	err := h.app.AddGroups(r.Context(), userId, *groups)
	switch {
	case errors.Is(err, app.ErrInvalidGroupName):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case errors.Is(err, app.ErrNoDefaultProfile):
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	case errors.Is(err, app.ErrGroupExists):
		http.Error(w, err.Error(), http.StatusConflict)
		return
	case err != nil:
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// @Summary		Remove groups
// @Tags		groups
// @Accept		json
// @Param		body	body	[]model.Group	true	"Groups"
// @Success	200
// @Failure	400	{string}	string	"Request body is empty"
// @Failure	404	{string}	string	"No default profile or group not found"
// @Failure	500
// @Security	BearerAuth
// @Router		/removegroups [delete]
func (h *handler) removeGroups(w http.ResponseWriter, r *http.Request) {
	userId, httpErr := extractUserId(r)
	if httpErr != nil {
		http.Error(w, httpErr.msg, httpErr.code)
		return
	}

	groups, httpErr := readBody[[]model.Group](r)
	if httpErr != nil {
		http.Error(w, httpErr.Error(), httpErr.code)
		return
	}
	if groups == nil || len(*groups) == 0 {
		http.Error(w, "Request body is empty", http.StatusBadRequest)
		return
	}

	err := h.app.RemoveGroups(r.Context(), userId, *groups)
	switch {
	case errors.Is(err, app.ErrNoDefaultProfile):
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	case errors.Is(err, app.ErrGroupNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	case err != nil:
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// @Summary		Rename group
// @Tags		groups
// @Accept		json
// @Param		body	body	model.RenameGroupBody	true	"Rename group body"
// @Success	200
// @Failure	400	{string}	string	"Malformed request body"
// @Failure	404	{string}	string	"No default profile or group not found"
// @Failure	500
// @Security	BearerAuth
// @Router		/renamegroup [patch]
func (h *handler) renameGroup(w http.ResponseWriter, r *http.Request) {
	userId, httpErr := extractUserId(r)
	if httpErr != nil {
		http.Error(w, httpErr.msg, httpErr.code)
		return
	}

	group, httpErr := readBody[model.RenameGroupBody](r)
	if httpErr != nil {
		http.Error(w, httpErr.Error(), httpErr.code)
		return
	}
	if group.NameNew == "" || group.NameOld == "" {
		http.Error(w, "Malformed request body", http.StatusBadRequest)
		return
	}

	err := h.app.RenameGroup(r.Context(), userId, *group)
	switch {
	case errors.Is(err, app.ErrNoDefaultProfile):
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	case errors.Is(err, app.ErrGroupNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	case err != nil:
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// @Summary		Update weights
// @Tags		groups
// @Accept		json
// @Param		body	body	[]model.Group	true	"Groups"
// @Success	200
// @Failure	400	{string}	string	"Request body is empty"
// @Failure	404	{string}	string	"No default profile or group not found"
// @Failure	500
// @Security	BearerAuth
// @Router		/updateweights [patch]
func (h *handler) updateWeights(w http.ResponseWriter, r *http.Request) {
	userId, httpErr := extractUserId(r)
	if httpErr != nil {
		http.Error(w, httpErr.msg, httpErr.code)
		return
	}

	groups, httpErr := readBody[[]model.Group](r)
	if httpErr != nil {
		http.Error(w, httpErr.Error(), httpErr.code)
		return
	}
	if groups == nil || len(*groups) == 0 {
		http.Error(w, "Request body is empty", http.StatusBadRequest)
		return
	}

	err := h.app.UpdateWeights(r.Context(), userId, *groups)
	switch {
	case errors.Is(err, app.ErrInvalidWeights):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case errors.Is(err, app.ErrNoDefaultProfile):
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	case errors.Is(err, app.ErrGroupNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	case err != nil:
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
