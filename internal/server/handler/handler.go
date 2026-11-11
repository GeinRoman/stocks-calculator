package handler

import (
	"context"
	"fmt"
	"net/http"
	"stocks_calculator/internal/model"
	"time"
)

type (
	App interface {
		Login(ctx context.Context, data model.AuthModel) (*model.AuthResponse, error)
		CreateUser(ctx context.Context, data model.AuthModel) (*model.AuthResponse, error)
		RefreshToken(ctx context.Context, refToken string) (*model.Token, error)

		RemoveProfile(ctx context.Context, profile model.Profile, userId int) error
		CreateProfile(ctx context.Context, profile model.Profile, userId int) error
		GetProfiles(ctx context.Context, userId int) ([]model.Profile, error)
		SetDefaultProfile(ctx context.Context, profile model.Profile, userId int) error

		GetGroups(ctx context.Context, userId int) ([]model.StockGroup, error)
		AddGroups(ctx context.Context, userId int, groups []model.StockGroup) error
		RemoveGroups(ctx context.Context, userId int, groups []model.StockGroup) error
		RenameGroup(ctx context.Context, userId int, group model.RenameStockGroupBody) error
		UpdateWeights(ctx context.Context, userId int, groups []model.StockGroup) error

		AddStocks(ctx context.Context, userId int, stocks []model.Stock) error
		RemoveStocks(ctx context.Context, userId int, stocks []model.Stock) error
		GetStocks(ctx context.Context, userId int) ([]model.Stock, error)
		FindStocks(ctx context.Context, userId int, names []string) ([]model.FoundStock, error)

		GetBondGroups(ctx context.Context, userId int) ([]model.BondGroup, error)
		AddBondGroups(ctx context.Context, userId int, groups []model.BondGroup) error
		RemoveBondGroups(ctx context.Context, userId int, groups []model.BondGroup) error
		RenameBondGroup(ctx context.Context, userId int, group model.RenameBondGroupBody) error
		UpdateBondWeights(ctx context.Context, userId int, groups []model.BondGroup) error

		AddBonds(ctx context.Context, userId int, bonds []model.BondOperation) error
		RemoveBonds(ctx context.Context, userId int, bonds []model.BondOperation) error
		GetBonds(ctx context.Context, userId int) ([]model.Bond, error)
		FindBonds(ctx context.Context, userId int, names []string) ([]model.FoundBond, error)
		GetBondCoupons(ctx context.Context, code string) (model.BondCouponsResponse, error)
		GetCurrencyRate(ctx context.Context, from, to string) (model.CurrencyRate, error)

		GetCryptoGroups(context.Context, int) ([]model.CryptoGroup, error)
		AddCryptoGroups(context.Context, int, []model.CryptoGroup) error
		RemoveCryptoGroups(context.Context, int, []model.CryptoGroup) error
		RenameCryptoGroup(context.Context, int, model.RenameCryptoGroupBody) error
		UpdateCryptoWeights(context.Context, int, []model.CryptoGroup) error
		GetCryptoAssets(context.Context, int) ([]model.CryptoAsset, error)
		UpdateCryptoAsset(context.Context, int, model.CryptoAssetUpdate) error
		BuyCrypto(context.Context, int, []model.CryptoOperation) error
		SellCrypto(context.Context, int, []model.CryptoOperation) error

		Rebalance(ctx context.Context, userId int, noSell bool, valueDiff float64) (model.StockRebalanceResponse, error)
		RebalanceBonds(ctx context.Context, userId int, noSell bool, valueDiff float64) (model.BondRebalanceResponse, error)
	}
	TokenManager interface {
		ExtractClaims(tokenStr string) (model.Claims, error)
	}

	handler struct {
		app App
		tm  TokenManager
	}
)

func NewHttpServer(app App, tm TokenManager, port int) *http.Server {
	h := handler{app: app, tm: tm}
	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,

		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		WriteTimeout:      40 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

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

	mux.HandleFunc("GET /getbondgroups", h.accessTokenValidation(h.getBondGroups))
	mux.HandleFunc("POST /addbondgroups", h.accessTokenValidation(h.addBondGroups))
	mux.HandleFunc("DELETE /removebondgroups", h.accessTokenValidation(h.removeBondGroups))
	mux.HandleFunc("PATCH /renamebondgroup", h.accessTokenValidation(h.renameBondGroup))
	mux.HandleFunc("PATCH /updatebondweights", h.accessTokenValidation(h.updateBondWeights))

	mux.HandleFunc("POST /addbonds", h.accessTokenValidation(h.addBonds))
	mux.HandleFunc("DELETE /removebonds", h.accessTokenValidation(h.removeBonds))
	mux.HandleFunc("GET /getbonds", h.accessTokenValidation(h.getBonds))
	mux.HandleFunc("POST /findbonds", h.accessTokenValidation(h.findBonds))
	mux.HandleFunc("GET /bondcoupons", h.accessTokenValidation(h.getBondCoupons))
	mux.HandleFunc("GET /currencyrate", h.accessTokenValidation(h.getCurrencyRate))

	mux.HandleFunc("GET /getcryptogroups", h.accessTokenValidation(h.getCryptoGroups))
	mux.HandleFunc("GET /getcryptoassets", h.accessTokenValidation(h.getCryptoAssets))
	mux.HandleFunc("POST /addcryptogroups", h.accessTokenValidation(h.addCryptoGroups))
	mux.HandleFunc("DELETE /removecryptogroups", h.accessTokenValidation(h.removeCryptoGroups))
	mux.HandleFunc("PATCH /renamecryptogroup", h.accessTokenValidation(h.renameCryptoGroup))
	mux.HandleFunc("PATCH /updatecryptoasset", h.accessTokenValidation(h.updateCryptoAsset))
	mux.HandleFunc("PATCH /updatecryptoweights", h.accessTokenValidation(h.updateCryptoWeights))
	mux.HandleFunc("POST /buycrypto", h.accessTokenValidation(h.buyCrypto))
	mux.HandleFunc("DELETE /sellcrypto", h.accessTokenValidation(h.sellCrypto))

	mux.HandleFunc("GET /rebalance", h.accessTokenValidation(h.rebalance))
	mux.HandleFunc("GET /rebalancebonds", h.accessTokenValidation(h.rebalanceBonds))

	registerSwagger(mux)
	return server
}

// @Summary      Test
// @Description  Grant access token for valid refresh token
// @Router       /test [get]
// @Security BearerAuth
func (h *handler) test(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
