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

		RemoveProfile(ctx context.Context, profile model.Profile, userId int) error
		CreateProfile(ctx context.Context, profile model.Profile, userId int, def bool) error
		GetProfiles(ctx context.Context, userId int) ([]model.Profile, error)
		SetDefaultProfile(ctx context.Context, profile model.Profile, userId int) error

		GetGroups(ctx context.Context, userId int) ([]model.Group, error)
		AddGroups(ctx context.Context, userId int, groups []model.Group) error
		RemoveGroups(ctx context.Context, userId int, groups []model.Group) error
		RenameGroup(ctx context.Context, userId int, group model.RenameGroupBody) error
		UpdateWeights(ctx context.Context, userId int, groups []model.Group) error

		AddStocks(ctx context.Context, userId int, stocks []model.Stock) error
		RemoveStocks(ctx context.Context, userId int, stocks []model.Stock) error
		GetStocks(ctx context.Context, userId int) ([]model.Stock, error)
		FindStocks(ctx context.Context, names []string) []model.FoundStock

		Rebalance(ctx context.Context, userId int, noSell bool, valueDiff float64) (model.RebalanceResponse, error)
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
	mux.HandleFunc("DELETE /removeprofile", h.accessTokenValidation(h.removeProfile))
	mux.HandleFunc("POST /createprofile", h.accessTokenValidation(h.createProfile))
	mux.HandleFunc("GET /getprofiles", h.accessTokenValidation(h.getProfiles))
	mux.HandleFunc("PATCH /setdefaultprofile", h.accessTokenValidation(h.setDefaultProfile))

	mux.HandleFunc("GET /getgroups", h.accessTokenValidation(h.getGroups))
	mux.HandleFunc("POST /addgroups", h.accessTokenValidation(h.addGroups))
	mux.HandleFunc("DELETE /removegroups", h.accessTokenValidation(h.removeGroups))
	mux.HandleFunc("PATCH /renamegroup", h.accessTokenValidation(h.renameGroup))
	mux.HandleFunc("PATCH /updateweights", h.accessTokenValidation(h.updateWeights))

	mux.HandleFunc("POST /addstocks", h.accessTokenValidation(h.addStocks))
	mux.HandleFunc("DELETE /removestocks", h.accessTokenValidation(h.removeStocks))
	mux.HandleFunc("GET /getstocks", h.accessTokenValidation(h.getStocks))
	mux.HandleFunc("POST /findstocks", h.accessTokenValidation(h.findStocks))

	mux.HandleFunc("GET /rebalance", h.accessTokenValidation(h.rebalance))

	registerSwagger(mux)
	return mux
}

// @Summary      Test
// @Description  Grant access token for valid refresh token
// @Router       /test [get]
// @Security BearerAuth
func (h *handler) test(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
