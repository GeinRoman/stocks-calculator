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
// @Param        logininfo  body      model.LoginModel  true  "Login credentials"
// @Success      200   {object}  model.LoginResponse
// @Failure      400   {string}  string  "incorrect request body format"
// @Failure      401   {string}  string  "wrong credentials"
// @Failure      409   {string}  string  "user exists"
// @Failure      500   {string}  string  "internal server error"
// @Router       /login [post]
func login(w http.ResponseWriter, r *http.Request) {
	data, httpErr := readBody[model.LoginModel](r)
	if httpErr != nil {
		http.Error(w, httpErr.Error(), httpErr.code)
		return
	}

	loginResponse, err := app.Login(*data)
	switch {
	case errors.Is(err, app.ErrUserExists):
		http.Error(w, err.Error(), http.StatusConflict)
		return
	case errors.Is(err, app.ErrWrongCredentials):
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeAndMarshal(w, loginResponse)
}
