package handler

import (
	"context"
	"net/http"
	"stocks_calculator/internal/model"
)

type (
	App interface {
		Login(ctx context.Context, data model.AuthModel) (*model.AuthResponse, error)
		CreateUser(ctx context.Context, data model.AuthModel) (*model.AuthResponse, error)
		RefreshToken(ctx context.Context, refToken string) (*model.Token, error)
	}
	TokenManager interface {
		ExtractClaims(tokenStr string) (model.Claims, error)
	}

	handler struct {
		app App
		tm  TokenManager
	}
)

func New(app App, tm TokenManager) *http.ServeMux {
	h := handler{app: app, tm: tm}
	mux := http.NewServeMux()
	// server := &http.Server{
	// 	Handler:      mux,
	// 	ReadTimeout:  10 * time.Second,
	// 	WriteTimeout: 10 * time.Second,
	// 	IdleTimeout:  120 * time.Second,
	// }

	//unauthorized endpoints
	mux.HandleFunc("POST /login", h.login)
	mux.HandleFunc("POST /createuser", h.createUser)
	mux.HandleFunc("POST /reftoken", h.refreshToken)

	//authorized endpoints
	mux.HandleFunc("GET /test", h.accessTokenValidation(h.test))

	registerSwagger(mux)
	return mux
}

// @Summary      Test
// @Description  Grant access token for valid refresh token
// @Router       /test [get]
// @Security BearerAuth
func (h *handler) test(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	return
}
