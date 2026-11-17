package handler

import (
	"context"
	"net/http"
	"testing"

	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/servererrors"
)

type cryptoHandlerApp struct {
	App
	user   int
	asset  model.CryptoAssetUpdate
	err    error
	called string
}

func (a *cryptoHandlerApp) AddCryptoGroups(_ context.Context, user int, _ []model.CryptoGroup) error {
	a.user, a.called = user, "add"
	return a.err
}
func (a *cryptoHandlerApp) RemoveCryptoGroups(_ context.Context, user int, _ []model.CryptoGroup) error {
	a.user, a.called = user, "remove"
	return a.err
}
func (a *cryptoHandlerApp) RenameCryptoGroup(_ context.Context, user int, _ model.RenameCryptoGroupBody) error {
	a.user, a.called = user, "rename"
	return a.err
}
func (a *cryptoHandlerApp) UpdateCryptoWeights(_ context.Context, user int, _ []model.CryptoGroup) error {
	a.user, a.called = user, "weights"
	return a.err
}
func (a *cryptoHandlerApp) BuyCrypto(_ context.Context, user int, _ []model.CryptoOperation) error {
	a.user, a.called = user, "buy"
	return a.err
}
func (a *cryptoHandlerApp) SellCrypto(_ context.Context, user int, _ []model.CryptoOperation) error {
	a.user, a.called = user, "sell"
	return a.err
}
func TestCryptoMutationRoutes(t *testing.T) {
	for _, tc := range []struct {
		method, path, body, called string
		list                       bool
	}{
		{"POST", "/addcryptogroups", `[{"name":"coins"}]`, "add", true},
		{"DELETE", "/removecryptogroups", `[{"name":"coins"}]`, "remove", true},
		{"PATCH", "/renamecryptogroup", `{}`, "rename", false},
		{"PATCH", "/updatecryptoweights", `[{"name":"coins"}]`, "weights", true},
		{"POST", "/buycrypto", `[{}]`, "buy", true},
		{"DELETE", "/sellcrypto", `[{}]`, "sell", true},
	} {
		t.Run(tc.path, func(t *testing.T) {
			a := &cryptoHandlerApp{}
			mux := NewHttpServer(a, bondTestTokens{}, 0).Handler
			w := bondRequest(mux, tc.method, tc.path, tc.body, "alice")
			if w.Code != http.StatusOK || a.called != tc.called || a.user == 0 {
				t.Fatalf("status %d called %q user %d", w.Code, a.called, a.user)
			}
			a.called = ""
			if tc.list {
				for _, body := range []string{"[]", "null"} {
					w = bondRequest(mux, tc.method, tc.path, body, "alice")
					if w.Code != http.StatusBadRequest || w.Body.String() != servererrors.ErrInvalidCrypto.Error()+"\n" || a.called != "" {
						t.Fatalf("body %q: status %d response %q called %q", body, w.Code, w.Body.String(), a.called)
					}
				}
			}
			a.err = servererrors.ErrCryptoNotFound
			w = bondRequest(mux, tc.method, tc.path, tc.body, "alice")
			if w.Code != http.StatusNotFound {
				t.Fatalf("error status %d", w.Code)
			}
		})
	}
}

func TestDeleteCryptoRouteRemoved(t *testing.T) {
	mux := NewHttpServer(&cryptoHandlerApp{}, bondTestTokens{}, 0).Handler
	if w := bondRequest(mux, http.MethodDelete, "/deletecrypto", `[{}]`, "alice"); w.Code != http.StatusNotFound {
		t.Fatalf("status %d, expected 404", w.Code)
	}
}

func (a *cryptoHandlerApp) UpdateCryptoAsset(_ context.Context, user int, asset model.CryptoAssetUpdate) error {
	a.user, a.asset = user, asset
	return a.err
}

func (a *cryptoHandlerApp) GetCryptoGroups(_ context.Context, user int) ([]model.CryptoGroup, error) {
	a.user = user
	return []model.CryptoGroup{}, a.err
}

func (a *cryptoHandlerApp) GetCryptoAssets(_ context.Context, user int) ([]model.CryptoAsset, error) {
	a.user = user
	return []model.CryptoAsset{}, a.err
}

func TestCryptoGetRoutesJSON(t *testing.T) {
	for _, path := range []string{"/getcryptogroups", "/getcryptoassets"} {
		t.Run(path, func(t *testing.T) {
			a := &cryptoHandlerApp{}
			mux := NewHttpServer(a, bondTestTokens{}, 0).Handler
			w := bondRequest(mux, http.MethodGet, path, "", "alice")
			if w.Code != http.StatusOK || w.Header().Get("Content-Type") != "application/json" || w.Body.String() != "[]" || a.user == 0 {
				t.Fatalf("status %d content-type %q body %q user %d", w.Code, w.Header().Get("Content-Type"), w.Body.String(), a.user)
			}
			a.err = servererrors.ErrCryptoNotFound
			w = bondRequest(mux, http.MethodGet, path, "", "alice")
			if w.Code != http.StatusNotFound || w.Header().Get("Content-Type") == "application/json" {
				t.Fatalf("error status %d content-type %q", w.Code, w.Header().Get("Content-Type"))
			}
		})
	}
}

func TestCryptoUpdateRoute(t *testing.T) {
	a := &cryptoHandlerApp{}
	mux := NewHttpServer(a, bondTestTokens{}, 0).Handler
	body := `{"id":42,"name":"BTC","platform":"wallet","url":"example","annual_interest":2}`
	if w := bondRequest(mux, http.MethodPatch, "/updatecryptoasset", body, ""); w.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated status %d", w.Code)
	}
	w := bondRequest(mux, http.MethodPatch, "/updatecryptoasset", body, "alice")
	if w.Code != http.StatusOK || a.user == 0 || a.asset.ID != 42 || a.asset.Platform != "wallet" {
		t.Fatalf("status %d user %d asset %+v", w.Code, a.user, a.asset)
	}
	a.err = servererrors.ErrCryptoExists
	if w := bondRequest(mux, http.MethodPatch, "/updatecryptoasset", body, "alice"); w.Code != http.StatusConflict {
		t.Fatalf("conflict status %d", w.Code)
	}
	a.err = servererrors.ErrCryptoNotFound
	if w := bondRequest(mux, http.MethodPatch, "/updatecryptoasset", body, "alice"); w.Code != http.StatusNotFound {
		t.Fatalf("not found status %d", w.Code)
	}
}
