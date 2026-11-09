package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/servererrors"
)

type bondTestApp struct {
	App
	called string
	user   int
	body   any
	err    error
}

func (a *bondTestApp) record(name string, user int, body any) error {
	a.called, a.user, a.body = name, user, body
	return a.err
}
func (a *bondTestApp) GetBondGroups(_ context.Context, user int) ([]model.BondGroup, error) {
	err := a.record("groups", user, nil)
	return []model.BondGroup{{Name: "Gov", Weight: 75}}, err
}
func (a *bondTestApp) AddBondGroups(_ context.Context, user int, v []model.BondGroup) error {
	return a.record("add groups", user, v)
}
func (a *bondTestApp) RemoveBondGroups(_ context.Context, user int, v []model.BondGroup) error {
	return a.record("remove groups", user, v)
}
func (a *bondTestApp) RenameBondGroup(_ context.Context, user int, v model.RenameBondGroupBody) error {
	return a.record("rename", user, v)
}
func (a *bondTestApp) UpdateBondWeights(_ context.Context, user int, v []model.BondGroup) error {
	return a.record("weights", user, v)
}
func (a *bondTestApp) AddBonds(_ context.Context, user int, v []model.Bond) error {
	return a.record("add bonds", user, v)
}
func (a *bondTestApp) RemoveBonds(_ context.Context, user int, v []model.Bond) error {
	return a.record("remove bonds", user, v)
}
func (a *bondTestApp) GetBonds(_ context.Context, user int) ([]model.Bond, error) {
	err := a.record("bonds", user, nil)
	return []model.Bond{{Code: "SU1", LotAmount: 2}}, err
}
func (a *bondTestApp) FindBonds(_ context.Context, user int, v []string) ([]model.FoundBond, error) {
	err := a.record("find", user, v)
	return []model.FoundBond{{SearchResults: []model.Bond{{Code: "SU1"}}, ErrMsg: ""}}, err
}

type bondTestTokens struct{ TokenManager }

func (bondTestTokens) ExtractClaims(token string) (model.Claims, error) {
	switch token {
	case "alice":
		return model.Claims{UserId: 11}, nil
	case "bob":
		return model.Claims{UserId: 22}, nil
	default:
		return model.Claims{}, errors.New("invalid token")
	}
}

type bondRoute struct {
	method, path, body, call string
	wantBody                 any
	wantJSON                 any
}

func bondRoutes() []bondRoute {
	groups := []model.BondGroup{{Name: "Gov", Weight: 75}}
	bonds := []model.Bond{{Code: "SU1", LotAmount: 2}}
	return []bondRoute{
		{"GET", "/getbondgroups", "", "groups", nil, groups},
		{"POST", "/addbondgroups", `[{"name":"Gov","weight":75}]`, "add groups", groups, nil},
		{"DELETE", "/removebondgroups", `[{"name":"Gov","weight":75}]`, "remove groups", groups, nil},
		{"PATCH", "/renamebondgroup", `{"name_old":"Gov","name_new":"Corp"}`, "rename", model.RenameBondGroupBody{NameOld: "Gov", NameNew: "Corp"}, nil},
		{"PATCH", "/updatebondweights", `[{"name":"Gov","weight":75}]`, "weights", groups, nil},
		{"POST", "/addbonds", `[{"code":"SU1","lot_amount":2}]`, "add bonds", bonds, nil},
		{"DELETE", "/removebonds", `[{"code":"SU1","lot_amount":2}]`, "remove bonds", bonds, nil},
		{"GET", "/getbonds", "", "bonds", nil, bonds},
		{"POST", "/findbonds", `["SU1"]`, "find", []string{"SU1"}, []model.FoundBond{{SearchResults: []model.Bond{{Code: "SU1"}}}}},
	}
}

func bondRequest(h http.Handler, method, path, body, token string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func TestBondRoutesDispatchAndAuthentication(t *testing.T) {
	for _, route := range bondRoutes() {
		t.Run(route.path, func(t *testing.T) {
			app := &bondTestApp{}
			h := NewHttpServer(app, bondTestTokens{}, 0).Handler
			for _, token := range []string{"", "invalid"} {
				w := bondRequest(h, route.method, route.path, route.body, token)
				if w.Code != http.StatusUnauthorized || app.called != "" {
					t.Fatalf("token %q: status %d, app call %q", token, w.Code, app.called)
				}
			}
			for token, user := range map[string]int{"alice": 11, "bob": 22} {
				w := bondRequest(h, route.method, route.path, route.body, token)
				if w.Code != http.StatusOK || app.called != route.call || app.user != user || !reflect.DeepEqual(app.body, route.wantBody) {
					t.Fatalf("token %q: status %d, call %q, user %d, body %#v", token, w.Code, app.called, app.user, app.body)
				}
				if route.wantJSON != nil {
					want, err := json.Marshal(route.wantJSON)
					if err != nil || w.Body.String() != string(want) {
						t.Fatalf("JSON = %s, want %s (marshal error %v)", w.Body.String(), want, err)
					}
				} else if w.Body.Len() != 0 {
					t.Fatalf("unexpected response body: %s", w.Body.String())
				}
			}
			w := bondRequest(h, "PUT", route.path, route.body, "alice")
			if w.Code != http.StatusMethodNotAllowed {
				t.Fatalf("wrong method: status %d", w.Code)
			}
		})
	}
}

func TestBondRoutesInvalidJSON(t *testing.T) {
	for _, route := range bondRoutes() {
		if route.body == "" {
			continue
		}
		t.Run(route.path, func(t *testing.T) {
			for _, body := range []string{"", "{", "null", "[]"} {
				app := &bondTestApp{}
				w := bondRequest(NewHttpServer(app, bondTestTokens{}, 0).Handler, route.method, route.path, body, "alice")
				if w.Code != http.StatusBadRequest || app.called != "" {
					t.Errorf("body %q: status %d, call %q", body, w.Code, app.called)
				}
			}
		})
	}
}

func TestBondRoutesAppErrors(t *testing.T) {
	for _, route := range bondRoutes() {
		t.Run(route.path, func(t *testing.T) {
			cases := []struct {
				err    error
				status int
			}{
				{servererrors.ErrNoDefaultProfile, http.StatusNotFound},
				{errors.New("internal failure"), http.StatusInternalServerError},
			}
			switch route.call {
			case "add groups":
				cases = append(cases, struct {
					err    error
					status int
				}{servererrors.ErrInvalidGroupName, 400}, struct {
					err    error
					status int
				}{servererrors.ErrGroupExists, 409})
			case "remove groups", "rename":
				cases = append(cases, struct {
					err    error
					status int
				}{servererrors.ErrGroupNotFound, 404})
			case "weights":
				cases = append(cases, struct {
					err    error
					status int
				}{servererrors.ErrInvalidWeights, 400}, struct {
					err    error
					status int
				}{servererrors.ErrGroupNotFound, 404})
			case "add bonds", "remove bonds":
				for _, err := range []error{servererrors.ErrWrongBondGroup, servererrors.ErrWrongBondAmount, servererrors.ErrDuplicateBond} {
					cases = append(cases, struct {
						err    error
						status int
					}{err, 400})
				}
				if route.call == "add bonds" {
					cases = append(cases, struct {
						err    error
						status int
					}{servererrors.ErrGroupNotFound, 404})
				} else {
					cases = append(cases, struct {
						err    error
						status int
					}{servererrors.ErrBondNotFound, 404})
				}
			}
			for _, tc := range cases {
				app := &bondTestApp{err: fmt.Errorf("wrapped: %w", tc.err)}
				w := bondRequest(NewHttpServer(app, bondTestTokens{}, 0).Handler, route.method, route.path, route.body, "alice")
				if w.Code != tc.status || app.called != route.call {
					t.Errorf("error %v: status %d, call %q", tc.err, w.Code, app.called)
				}
				if tc.status != http.StatusInternalServerError && w.Body.String() != app.err.Error()+"\n" {
					t.Errorf("error %v: body %q", tc.err, w.Body.String())
				}
			}
		})
	}
}
