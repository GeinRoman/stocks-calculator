package handler

import (
	"context"
	"net/http"
	"stocks_calculator/internal/model"
)

type App interface {
	Login(ctx context.Context, data model.AuthModel) (*model.AuthResponse, error)
	CreateUser(ctx context.Context, data model.AuthModel) (*model.AuthResponse, error)
	RefreshToken(ctx context.Context, refToken string) (*model.Token, error)
}

type handler struct {
	app App
}

func New(app App) *http.ServeMux {
	h := handler{app: app}
	mux := http.NewServeMux()
	// server := &http.Server{
	// 	Handler:      mux,
	// 	ReadTimeout:  10 * time.Second,
	// 	WriteTimeout: 10 * time.Second,
	// 	IdleTimeout:  120 * time.Second,
	// }

	mux.HandleFunc("POST /login", h.login)
	mux.HandleFunc("POST /createuser", h.createUser)
	mux.HandleFunc("POST /reftoken", h.refreshToken)

	registerSwagger(mux)
	return mux
}
