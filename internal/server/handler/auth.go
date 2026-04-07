package handler

import (
	"errors"
	"net/http"
	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/app"
)

// @Summary      Login
// @Description  Log in with username and password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        authmodel  body      model.AuthModel  true  "Login credentials"
// @Success      200   {object}  model.AuthResponse
// @Failure      400   {string}  string  "incorrect request body format"
// @Failure      401   {string}  string  "wrong credentials"
// @Failure      500   {string}  string  "internal server error"
// @Router       /login [post]
func (h *handler) login(w http.ResponseWriter, r *http.Request) {
	data, httpErr := readBody[model.AuthModel](r)
	if httpErr != nil {
		http.Error(w, httpErr.Error(), httpErr.code)
		return
	}
	if data.User == "" || data.Pass == "" || len(data.Pass) > 64 {
		http.Error(w, "Empty credentials or password length exceeds 64", http.StatusBadRequest)
		return
	}

	authResponse, err := h.app.Login(r.Context(), *data)
	switch {
	case errors.Is(err, app.ErrWrongCredentials):
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeAndMarshal(w, authResponse)
}

// @Summary      CreateUser
// @Description  Create user with username and password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        authmodel  body      model.AuthModel  true  "Create user credentials"
// @Success      200   {object}  model.AuthResponse
// @Failure      400   {string}  string  "incorrect request body format"
// @Failure      409   {string}  string  "user exists"
// @Failure      500   {string}  string  "internal server error"
// @Router       /createuser [post]
func (h *handler) createUser(w http.ResponseWriter, r *http.Request) {
	data, httpErr := readBody[model.AuthModel](r)
	if httpErr != nil {
		http.Error(w, httpErr.Error(), httpErr.code)
		return
	}
	if data.User == "" || data.Pass == "" || len(data.Pass) > 64 {
		http.Error(w, "Empty credentials or password length exceeds 64", http.StatusBadRequest)
		return
	}

	authResponse, err := h.app.CreateUser(r.Context(), *data)
	switch {
	case errors.Is(err, app.ErrUserExists):
		http.Error(w, err.Error(), http.StatusConflict)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeAndMarshal(w, authResponse)
}

// @Summary      RefreshToken
// @Description  Grant access token for valid refresh token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        tokenmodel  body      string  true  "Refresh token"
// @Success      200   {object}  model.Token
// @Failure      400   {string}  string  "no refresh token provided"
// @Failure      404   {string}  string  "refresh token is invalid"
// @Failure      500   {string}  string  "internal server error"
// @Router       /reftoken [post]
func (h *handler) refreshToken(w http.ResponseWriter, r *http.Request) {
	data, httpErr := readBody[string](r)
	if httpErr != nil {
		http.Error(w, httpErr.Error(), httpErr.code)
		return
	}

	refToken := *data
	if refToken == "" {
		http.Error(w, "No refresh token provided", http.StatusBadRequest)
		return
	}
	if len(refToken) != 64 {
		http.Error(w, app.ErrInvalidRefreshToken.Error(), http.StatusNotFound)
		return
	}

	token, err := h.app.RefreshToken(r.Context(), refToken)
	switch {
	case errors.Is(err, app.ErrInvalidRefreshToken):
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeAndMarshal(w, token)
}
