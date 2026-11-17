package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/servererrors"
)

type transactionHistoryTestApp struct {
	App
	user   int
	called string
	cursor model.TransactionCursor
	err    error
}

func (a *transactionHistoryTestApp) record(name string, user int, c model.TransactionCursor) {
	a.called, a.user, a.cursor = name, user, c
}
func (a *transactionHistoryTestApp) GetDepositTransactions(_ context.Context, u int, c model.TransactionCursor) ([]model.DepositTransaction, error) {
	a.record("deposit", u, c)
	return []model.DepositTransaction{{ID: 10}}, a.err
}
func (a *transactionHistoryTestApp) GetInsuranceTransactions(_ context.Context, u int, c model.TransactionCursor) ([]model.InsuranceTransaction, error) {
	a.record("insurance", u, c)
	return []model.InsuranceTransaction{{ID: 10}}, a.err
}
func (a *transactionHistoryTestApp) GetCryptoTransactions(_ context.Context, u int, c model.TransactionCursor) ([]model.CryptoTransaction, error) {
	a.record("crypto", u, c)
	return []model.CryptoTransaction{{ID: 10}}, a.err
}
func (a *transactionHistoryTestApp) GetStockTransactions(_ context.Context, u int, c model.TransactionCursor) ([]model.StockTransaction, error) {
	a.record("stock", u, c)
	return []model.StockTransaction{{ID: 10}}, a.err
}
func (a *transactionHistoryTestApp) GetBondTransactions(_ context.Context, u int, c model.TransactionCursor) ([]model.BondTransaction, error) {
	a.record("bond", u, c)
	return []model.BondTransaction{{ID: 10}}, a.err
}

func TestTransactionHistoryRoutes(t *testing.T) {
	for _, tc := range []struct{ path, name string }{
		{"/getdeposittransactions", "deposit"}, {"/getinsurancetransactions", "insurance"},
		{"/getcryptotransactions", "crypto"}, {"/getstocktransactions", "stock"}, {"/getbondtransactions", "bond"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := &transactionHistoryTestApp{}
			h := NewHttpServer(a, bondTestTokens{}, 0).Handler
			w := bondRequest(h, http.MethodGet, tc.path, "", "")
			require.Equal(t, http.StatusUnauthorized, w.Code)
			require.Empty(t, a.called)
			w = bondRequest(h, http.MethodGet, tc.path+"?before=2026-11-14T12%3A00%3A00.123456Z&before_id=42", "", "alice")
			require.Equal(t, http.StatusOK, w.Code)
			require.Equal(t, "application/json", w.Header().Get("Content-Type"))
			require.Equal(t, tc.name, a.called)
			require.Equal(t, 11, a.user)
			require.Equal(t, int64(42), *a.cursor.BeforeID)
			require.Equal(t, time.Date(2026, 11, 14, 12, 0, 0, 123456000, time.UTC), *a.cursor.Before)
			for _, query := range []string{"?before=", "?before=2026-01-01", "?before_id=1", "?before=2026-01-01T00:00:00Z&before_id=0", "?before=2026-01-01T00:00:00Z&before_id=x", "?before=x&before=y"} {
				a.called = ""
				w = bondRequest(h, http.MethodGet, tc.path+query, "", "alice")
				require.Equal(t, http.StatusBadRequest, w.Code, query)
				require.Empty(t, a.called)
			}
			for _, e := range []struct {
				err    error
				status int
			}{{servererrors.ErrNoDefaultProfile, 404}, {errors.New("error"), 500}} {
				a.err = e.err
				w = bondRequest(h, http.MethodGet, tc.path, "", "bob")
				require.Equal(t, e.status, w.Code)
				require.Equal(t, 22, a.user)
				require.Nil(t, a.cursor.Before)
			}
		})
	}
}
