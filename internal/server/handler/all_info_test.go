package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/servererrors"
)

type allInfoTestApp struct {
	App
	users []int
	fail  string
	err   error
	empty bool
}

func (a *allInfoTestApp) record(name string, user int) error {
	a.users = append(a.users, user)
	if a.fail == name {
		return a.err
	}
	return nil
}
func (a *allInfoTestApp) PersonalAssetProfileContext(ctx context.Context, user int) (context.Context, error) {
	return context.WithValue(ctx, testProfileKey{}, user), a.record("profile context", user)
}
func (a *allInfoTestApp) GetDeposits(ctx context.Context, user int) ([]model.Deposit, error) {
	if ctx.Value(testProfileKey{}) != user {
		panic("missing profile context")
	}
	err := a.record("deposits", user)
	if a.empty {
		return nil, err
	}
	return []model.Deposit{{Name: "deposit", StartDate: "2020-01-01"}}, err
}
func (a *allInfoTestApp) GetCreditCards(ctx context.Context, user int) ([]model.CreditCard, error) {
	if ctx.Value(testProfileKey{}) != user {
		panic("missing profile context")
	}
	err := a.record("cards", user)
	if a.empty {
		return nil, err
	}
	return []model.CreditCard{{Name: "card"}}, err
}
func (a *allInfoTestApp) GetInsurances(ctx context.Context, user int) ([]model.Insurance, error) {
	if ctx.Value(testProfileKey{}) != user {
		panic("missing profile context")
	}
	err := a.record("insurances", user)
	if a.empty {
		return nil, err
	}
	return []model.Insurance{{Name: "insurance", StartDate: "2020-01-01"}}, err
}
func (a *allInfoTestApp) GetBonds(_ context.Context, user int) ([]model.Bond, error) {
	err := a.record("bonds", user)
	if a.empty {
		return nil, err
	}
	return []model.Bond{{Code: "bond"}}, err
}
func (a *allInfoTestApp) GetCryptoAssets(_ context.Context, user int) ([]model.CryptoAsset, error) {
	err := a.record("crypto", user)
	if a.empty {
		return nil, err
	}
	return []model.CryptoAsset{{Name: "crypto"}}, err
}
func (a *allInfoTestApp) GetStocks(_ context.Context, user int) ([]model.Stock, error) {
	err := a.record("stocks", user)
	if a.empty {
		return nil, err
	}
	return []model.Stock{{Code: "stock"}}, err
}
func (a *allInfoTestApp) GetProfiles(_ context.Context, user int) ([]model.Profile, error) {
	err := a.record("profiles", user)
	if a.empty {
		return nil, err
	}
	return []model.Profile{{Name: "default", Default: true}}, err
}

func TestGetAll(t *testing.T) {
	a := &allInfoTestApp{}
	h := NewHttpServer(a, bondTestTokens{}, 0).Handler
	w := bondRequest(h, http.MethodGet, "/getall", "", "")
	require.Equal(t, http.StatusUnauthorized, w.Code)
	require.Empty(t, a.users)
	for _, tc := range []struct {
		token string
		user  int
	}{{"alice", 11}, {"bob", 22}} {
		a.users = nil
		w = bondRequest(h, http.MethodGet, "/getall", "", tc.token)
		require.Equal(t, http.StatusOK, w.Code)
		require.Equal(t, "application/json", w.Header().Get("Content-Type"))
		require.Equal(t, []int{tc.user, tc.user, tc.user, tc.user, tc.user, tc.user, tc.user, tc.user}, a.users)
		var got model.AllInfo
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
		require.Equal(t, "bond", got.Bonds[0].Code)
		require.Equal(t, "deposit", got.PersonalAssets.Deposits[0].Name)
		require.Equal(t, "2020-01-01", got.PersonalAssets.Deposits[0].StartDate)
		require.Equal(t, "card", got.PersonalAssets.CreditCards[0].Name)
		require.Equal(t, "insurance", got.PersonalAssets.Insurances[0].Name)
		require.Equal(t, "crypto", got.CryptoAssets[0].Name)
		require.Equal(t, "stock", got.Stocks[0].Code)
		require.Equal(t, "default", got.Profiles[0].Name)
	}
	a.empty = true
	w = bondRequest(h, http.MethodGet, "/getall", "", "alice")
	require.JSONEq(t, `{"bonds":[],"personal_assets":{"deposits":[],"credit_cards":[],"insurances":[]},"crypto_assets":[],"stocks":[],"profiles":[]}`, w.Body.String())
}

func TestGetAllErrors(t *testing.T) {
	for _, stage := range []string{"profile context", "deposits", "cards", "insurances", "bonds", "crypto", "stocks", "profiles"} {
		for _, tc := range []struct {
			err    error
			status int
		}{{servererrors.ErrNoDefaultProfile, http.StatusNotFound}, {errors.New("failure"), http.StatusInternalServerError}} {
			t.Run(stage+tc.err.Error(), func(t *testing.T) {
				a := &allInfoTestApp{fail: stage, err: tc.err}
				h := NewHttpServer(a, bondTestTokens{}, 0).Handler
				w := bondRequest(h, http.MethodGet, "/getall", "", "alice")
				require.Equal(t, tc.status, w.Code)
			})
		}
	}
}
