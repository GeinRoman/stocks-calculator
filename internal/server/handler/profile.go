package handler

import (
	"errors"
	"net/http"
	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/servererrors"
	"strconv"
)

// @Summary		Remove profile
// @Tags		profiles
// @Accept		json
// @Param		body	body	model.Profile	true	"Profile"
// @Success	200
// @Failure	400	{string}	string	"Missing profile name"
// @Failure	404	{string}	string	"Profile not found"
// @Failure	500
// @Security	BearerAuth
// @Router		/removeprofile [delete]
func (h *handler) removeProfile(w http.ResponseWriter, r *http.Request) {
	userId, httpErr := extractUserId(r)
	if httpErr != nil {
		http.Error(w, httpErr.msg, httpErr.code)
		return
	}

	profile, httpErr := readBody[model.Profile](r)
	if httpErr != nil {
		http.Error(w, httpErr.Error(), httpErr.code)
		return
	}
	if profile.Name == "" {
		http.Error(w, "Missing profile name", http.StatusBadRequest)
		return
	}

	err := h.app.RemoveProfile(r.Context(), *profile, userId)
	switch {
	case errors.Is(err, servererrors.ErrProfileNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	case err != nil:
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// @Summary		Create profile
// @Tags		profiles
// @Accept		json
// @Param		body	body	model.Profile	true	"Profile"
// @Success	200
// @Failure	400	{string}	string	"Missing profile name"
// @Failure	409	{string}	string	"Profile already exists"
// @Failure	500
// @Security	BearerAuth
// @Router		/createprofile [post]
func (h *handler) createProfile(w http.ResponseWriter, r *http.Request) {
	userId, httpErr := extractUserId(r)
	if httpErr != nil {
		http.Error(w, httpErr.msg, httpErr.code)
		return
	}

	profile, httpErr := readBody[model.Profile](r)
	if httpErr != nil {
		http.Error(w, httpErr.Error(), httpErr.code)
		return
	}
	if profile.Name == "" {
		http.Error(w, "Missing profile name", http.StatusBadRequest)
		return
	}

	err := h.app.CreateProfile(r.Context(), *profile, userId)
	switch {
	case errors.Is(err, servererrors.ErrProfileExists):
		http.Error(w, err.Error(), http.StatusConflict)
		return
	case err != nil:
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// @Summary		Get profiles
// @Tags		profiles
// @Produce	json
// @Success	200	{array}		model.Profile
// @Failure	500
// @Security	BearerAuth
// @Router		/getprofiles [get]
func (h *handler) getProfiles(w http.ResponseWriter, r *http.Request) {
	userId, httpErr := extractUserId(r)
	if httpErr != nil {
		http.Error(w, httpErr.msg, httpErr.code)
		return
	}

	profiles, err := h.app.GetProfiles(r.Context(), userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	writeAndMarshal(w, profiles)
}

// @Summary		Set default profile
// @Tags		profiles
// @Accept		json
// @Param		body	body	model.Profile	true	"Profile"
// @Success	200
// @Failure	400	{string}	string	"Missing profile name"
// @Failure	500
// @Security	BearerAuth
// @Router		/setdefaultprofile [patch]
func (h *handler) setDefaultProfile(w http.ResponseWriter, r *http.Request) {
	userId, httpErr := extractUserId(r)
	if httpErr != nil {
		http.Error(w, httpErr.msg, httpErr.code)
		return
	}

	profile, httpErr := readBody[model.Profile](r)
	if httpErr != nil {
		http.Error(w, httpErr.Error(), httpErr.code)
		return
	}
	if profile.Name == "" {
		http.Error(w, "Missing profile name", http.StatusBadRequest)
		return
	}

	err := h.app.SetDefaultProfile(r.Context(), *profile, userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
