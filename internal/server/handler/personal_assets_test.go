package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/servererrors"
)

type personalAssetsTestApp struct {
	App
	calls     []string
	user      int
	removed   int64
	err       error
	empty     bool
	deposit   model.DepositUpdate
	insurance model.InsuranceUpdate
	card      model.CreditCardUpdate
}

func (a *personalAssetsTestApp) UpdateDeposit(_ context.Context, user int, d model.DepositUpdate) error {
	a.user, a.deposit = user, d
	return a.err
}
func (a *personalAssetsTestApp) UpdateInsurance(_ context.Context, user int, d model.InsuranceUpdate) error {
	a.user, a.insurance = user, d
	return a.err
}
func (a *personalAssetsTestApp) UpdateCreditCard(_ context.Context, user int, c model.CreditCardUpdate) error {
	a.user, a.card = user, c
	return a.err
}

type testProfileKey struct{}

func (a *personalAssetsTestApp) PersonalAssetProfileContext(ctx context.Context, user int) (context.Context, error) {
	a.calls = append(a.calls, "profile")
	a.user = user
	return context.WithValue(ctx, testProfileKey{}, user), a.err
}

func (a *personalAssetsTestApp) GetDeposits(ctx context.Context, user int) ([]model.Deposit, error) {
	if ctx.Value(testProfileKey{}) != user {
		panic("missing resolved profile")
	}
	a.calls = append(a.calls, "deposits")
	a.user = user
	if a.empty {
		return nil, a.err
	}
	return []model.Deposit{{}}, a.err
}
func (a *personalAssetsTestApp) GetCreditCards(ctx context.Context, user int) ([]model.CreditCard, error) {
	if ctx.Value(testProfileKey{}) != user {
		panic("missing resolved profile")
	}
	a.calls = append(a.calls, "cards")
	a.user = user
	if a.empty {
		return nil, a.err
	}
	return []model.CreditCard{{}}, a.err
}
func (a *personalAssetsTestApp) GetInsurances(ctx context.Context, user int) ([]model.Insurance, error) {
	if ctx.Value(testProfileKey{}) != user {
		panic("missing resolved profile")
	}
	a.calls = append(a.calls, "insurances")
	a.user = user
	if a.empty {
		return nil, a.err
	}
	return []model.Insurance{{}}, a.err
}
func (a *personalAssetsTestApp) RemoveDeposit(_ context.Context, user int, id int64) error {
	a.user, a.removed = user, id
	return a.err
}
func (a *personalAssetsTestApp) RemoveInsurance(_ context.Context, user int, id int64) error {
	a.user, a.removed = user, id
	return a.err
}
func (a *personalAssetsTestApp) RemoveCreditCard(_ context.Context, user int, id int64) error {
	a.user, a.removed = user, id
	return a.err
}

func TestGetPersonalAssets(t *testing.T) {
	a := &personalAssetsTestApp{}
	h := handler{app: a}
	r := httptest.NewRequest(http.MethodGet, "/getpersonalassets", nil)
	r = r.WithContext(context.WithValue(r.Context(), userIdContextKey, 42))
	w := httptest.NewRecorder()
	h.getPersonalAssets(w, r)
	var got personalAssetsResponse
	if w.Code != http.StatusOK || json.Unmarshal(w.Body.Bytes(), &got) != nil || len(got.Deposits) != 1 || len(got.CreditCards) != 1 || len(got.Insurances) != 1 || strings.Join(a.calls, ",") != "profile,deposits,cards,insurances" || a.user != 42 {
		t.Fatalf("status %d, body %s, calls %v, user %d", w.Code, w.Body, a.calls, a.user)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(w.Body.Bytes(), &fields); err != nil || len(fields) != 3 || fields["deposits"] == nil || fields["credit_cards"] == nil || fields["insurances"] == nil {
		t.Fatalf("unexpected response keys: %s", w.Body)
	}
	a.empty = true
	w = httptest.NewRecorder()
	h.getPersonalAssets(w, r)
	if w.Code != http.StatusOK || w.Body.String() != `{"deposits":[],"credit_cards":[],"insurances":[]}` {
		t.Fatalf("empty assets: status %d, body %s", w.Code, w.Body)
	}
	a.calls = nil
	a.err = servererrors.ErrNoDefaultProfile
	w = httptest.NewRecorder()
	h.getPersonalAssets(w, r)
	if w.Code != http.StatusNotFound || len(a.calls) != 1 {
		t.Fatalf("status %d, calls %v", w.Code, a.calls)
	}
}

func TestPersonalAssetErrors(t *testing.T) {
	for _, tc := range []struct {
		err    error
		status int
	}{
		{model.ErrInvalidCreditCard, http.StatusBadRequest},
		{servererrors.ErrInvalidContract, http.StatusBadRequest},
		{model.ErrCreditCardNotFound, http.StatusNotFound},
		{servererrors.ErrNoDefaultProfile, http.StatusNotFound},
		{sql.ErrNoRows, http.StatusNotFound},
		{errors.New("unexpected"), http.StatusInternalServerError},
	} {
		w := httptest.NewRecorder()
		personalAssetError(w, fmt.Errorf("wrapped: %w", tc.err))
		if w.Code != tc.status {
			t.Errorf("%v: got status %d, want %d", tc.err, w.Code, tc.status)
		}
	}
}

func TestPersonalAssetsRequireAuthentication(t *testing.T) {
	a := &personalAssetsTestApp{}
	h := handler{app: a}
	for _, method := range []http.HandlerFunc{
		h.getPersonalAssets,
		h.addDeposit, h.updateDeposit, h.removeDeposit, h.closeDeposit,
		h.addInsurance, h.updateInsurance, h.removeInsurance, h.closeInsurance,
		h.addCreditCard, h.updateCreditCard, h.removeCreditCard,
	} {
		w := httptest.NewRecorder()
		h.accessTokenValidation(method)(w, httptest.NewRequest(http.MethodGet, "/", nil))
		if w.Code != http.StatusUnauthorized || len(a.calls) != 0 {
			t.Fatalf("status %d, app calls %v", w.Code, a.calls)
		}
	}
}

func TestRemovePersonalAssets(t *testing.T) {
	a := &personalAssetsTestApp{}
	h := handler{app: a}
	for _, method := range []http.HandlerFunc{h.removeDeposit, h.removeInsurance, h.removeCreditCard} {
		a.user, a.removed, a.err = 0, 0, nil
		r := httptest.NewRequest(http.MethodDelete, "/", strings.NewReader(`{"id":123}`))
		r = r.WithContext(context.WithValue(r.Context(), userIdContextKey, 42))
		w := httptest.NewRecorder()
		method(w, r)
		if w.Code != http.StatusOK || a.user != 42 || a.removed != 123 {
			t.Fatalf("status %d, user %d, id %d", w.Code, a.user, a.removed)
		}
		a.err = errors.New("unexpected")
		r = httptest.NewRequest(http.MethodDelete, "/", strings.NewReader(`{"id":123}`)).WithContext(r.Context())
		w = httptest.NewRecorder()
		method(w, r)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("status %d", w.Code)
		}
		r = httptest.NewRequest(http.MethodDelete, "/", strings.NewReader(`{`)).WithContext(r.Context())
		w = httptest.NewRecorder()
		method(w, r)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status %d", w.Code)
		}
	}
}

func TestPartialPersonalAssetBodies(t *testing.T) {
	a := &personalAssetsTestApp{}
	h := handler{app: a}
	for _, tc := range []struct {
		name   string
		method http.HandlerFunc
		body   string
		check  func() bool
	}{
		{"deposit omitted rate", h.updateDeposit, `{"id":1,"name":"new"}`, func() bool {
			return a.deposit.ID == 1 && a.deposit.Name != nil && *a.deposit.Name == "new" && a.deposit.AnnualInterest == nil
		}},
		{"deposit zero rate", h.updateDeposit, `{"id":1,"annual_interest":0}`, func() bool {
			return a.deposit.AnnualInterest != nil && *a.deposit.AnnualInterest == 0 && a.deposit.Name == nil
		}},
		{"insurance omitted rate", h.updateInsurance, `{"id":2,"name":"new"}`, func() bool {
			return a.insurance.ID == 2 && a.insurance.Name != nil && a.insurance.AnnualInterest == nil
		}},
		{"insurance zero rate", h.updateInsurance, `{"id":2,"annual_interest":0}`, func() bool {
			return a.insurance.AnnualInterest != nil && *a.insurance.AnnualInterest == 0 && a.insurance.Name == nil
		}},
		{"card omitted amounts", h.updateCreditCard, `{"id":3,"name":"new"}`, func() bool {
			return a.card.ID == 3 && a.card.Name != nil && a.card.MandatoryPaymentAmountRUB == nil && a.card.OverallLimitRUB == nil
		}},
		{"card zero amounts", h.updateCreditCard, `{"id":3,"mandatory_payment_amount_rub":0,"overall_limit_rub":0}`, func() bool {
			return a.card.Name == nil && a.card.MandatoryPaymentAmountRUB != nil && *a.card.MandatoryPaymentAmountRUB == 0 && a.card.OverallLimitRUB != nil && *a.card.OverallLimitRUB == 0
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(tc.body))
			r = r.WithContext(context.WithValue(r.Context(), userIdContextKey, 42))
			w := httptest.NewRecorder()
			tc.method(w, r)
			if w.Code != http.StatusOK || a.user != 42 || !tc.check() {
				t.Fatalf("status %d, user %d, body %s", w.Code, a.user, w.Body)
			}
		})
	}
}
