package handler

import (
	"context"
	"errors"
	"net/http"
	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/servererrors"
	"time"
)

const bondSearchTimeout = 35 * time.Second

// @Summary Add bonds
// @Tags bonds
// @Accept json
// @Param body body []model.BondOperation true "Bond purchases"
// @Success 200
// @Failure 400 {string} string "Request body is empty / invalid bond name, amount or price / wrong bond group / duplicate bond"
// @Failure 404 {string} string "No default profile or group not found"
// @Failure 500
// @Security BearerAuth
// @Router /addbonds [post]
func (h *handler) addBonds(w http.ResponseWriter, r *http.Request) {
	userId, httpErr := extractUserId(r)
	if httpErr != nil {
		http.Error(w, httpErr.msg, httpErr.code)
		return
	}
	bonds, httpErr := readBody[[]model.BondOperation](r)
	if httpErr != nil {
		http.Error(w, httpErr.Error(), httpErr.code)
		return
	}
	if bonds == nil || len(*bonds) == 0 {
		http.Error(w, "Request body is empty", http.StatusBadRequest)
		return
	}
	err := h.app.AddBonds(r.Context(), userId, *bonds)
	switch {
	case errors.Is(err, servererrors.ErrNoDefaultProfile), errors.Is(err, servererrors.ErrGroupNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, servererrors.ErrWrongBondGroup), errors.Is(err, servererrors.ErrWrongBondAmount), errors.Is(err, servererrors.ErrDuplicateBond), errors.Is(err, servererrors.ErrInvalidBondName), errors.Is(err, servererrors.ErrInvalidBondPrice):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case err != nil:
		w.WriteHeader(http.StatusInternalServerError)
	default:
		w.WriteHeader(http.StatusOK)
	}
}

// @Summary Remove bonds
// @Tags bonds
// @Accept json
// @Param body body []model.BondOperation true "Bond sales"
// @Success 200
// @Failure 400 {string} string "Request body is empty / invalid bond name, amount or price / wrong bond group / duplicate bond"
// @Failure 404 {string} string "No default profile or bond not found"
// @Failure 500
// @Security BearerAuth
// @Router /removebonds [delete]
func (h *handler) removeBonds(w http.ResponseWriter, r *http.Request) {
	userId, httpErr := extractUserId(r)
	if httpErr != nil {
		http.Error(w, httpErr.msg, httpErr.code)
		return
	}
	bonds, httpErr := readBody[[]model.BondOperation](r)
	if httpErr != nil {
		http.Error(w, httpErr.Error(), httpErr.code)
		return
	}
	if bonds == nil || len(*bonds) == 0 {
		http.Error(w, "Request body is empty", http.StatusBadRequest)
		return
	}
	err := h.app.RemoveBonds(r.Context(), userId, *bonds)
	switch {
	case errors.Is(err, servererrors.ErrNoDefaultProfile), errors.Is(err, servererrors.ErrBondNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, servererrors.ErrWrongBondGroup), errors.Is(err, servererrors.ErrWrongBondAmount), errors.Is(err, servererrors.ErrDuplicateBond), errors.Is(err, servererrors.ErrInvalidBondName), errors.Is(err, servererrors.ErrInvalidBondPrice):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case err != nil:
		w.WriteHeader(http.StatusInternalServerError)
	default:
		w.WriteHeader(http.StatusOK)
	}
}

// @Summary Get bonds
// @Tags bonds
// @Produce json
// @Success 200 {array} model.Bond
// @Failure 404 {string} string "No default profile"
// @Failure 500
// @Security BearerAuth
// @Router /getbonds [get]
func (h *handler) getBonds(w http.ResponseWriter, r *http.Request) {
	userId, httpErr := extractUserId(r)
	if httpErr != nil {
		http.Error(w, httpErr.msg, httpErr.code)
		return
	}
	bonds, err := h.app.GetBonds(r.Context(), userId)
	switch {
	case errors.Is(err, servererrors.ErrNoDefaultProfile):
		http.Error(w, err.Error(), http.StatusNotFound)
	case err != nil:
		w.WriteHeader(http.StatusInternalServerError)
	default:
		writeAndMarshal(w, bonds)
	}
}

// @Summary Find bonds
// @Tags bonds
// @Accept json
// @Param body body []string true "Bond names"
// @Produce json
// @Success 200 {array} model.FoundBond
// @Failure 400 {string} string "Request body is empty"
// @Failure 404 {string} string "No default profile"
// @Failure 500
// @Failure 504 {string} string "Bond search timed out"
// @Security BearerAuth
// @Router /findbonds [post]
func (h *handler) findBonds(w http.ResponseWriter, r *http.Request) {
	userId, httpErr := extractUserId(r)
	if httpErr != nil {
		http.Error(w, httpErr.msg, httpErr.code)
		return
	}
	names, httpErr := readBody[[]string](r)
	if httpErr != nil {
		http.Error(w, httpErr.Error(), httpErr.code)
		return
	}
	if names == nil || len(*names) == 0 {
		http.Error(w, "Request body is empty", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), bondSearchTimeout)
	defer cancel()
	bonds, err := h.app.FindBonds(ctx, userId, *names)
	switch {
	case errors.Is(ctx.Err(), context.DeadlineExceeded), errors.Is(err, context.DeadlineExceeded):
		http.Error(w, "Bond search timed out", http.StatusGatewayTimeout)
	case errors.Is(ctx.Err(), context.Canceled), errors.Is(err, context.Canceled):
		http.Error(w, "Bond search canceled", http.StatusRequestTimeout)
	case errors.Is(err, servererrors.ErrNoDefaultProfile):
		http.Error(w, err.Error(), http.StatusNotFound)
	case err != nil:
		w.WriteHeader(http.StatusInternalServerError)
	default:
		writeAndMarshal(w, bonds)
	}
}
