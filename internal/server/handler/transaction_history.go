package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"stocks_calculator/internal/model"
)

func transactionCursor(r *http.Request) (model.TransactionCursor, bool) {
	var cursor model.TransactionCursor
	query := r.URL.Query()
	if values, ok := query["before"]; ok {
		if len(values) != 1 {
			return cursor, false
		}
		before, err := time.Parse(time.RFC3339Nano, values[0])
		if err != nil {
			return cursor, false
		}
		cursor.Before = &before
	}
	if values, ok := query["before_id"]; ok {
		if len(values) != 1 || cursor.Before == nil {
			return cursor, false
		}
		id, err := strconv.ParseInt(values[0], 10, 64)
		if err != nil || id <= 0 {
			return cursor, false
		}
		cursor.BeforeID = &id
	}
	return cursor, true
}

func serveTransactionHistory[T any](w http.ResponseWriter, r *http.Request, get func(context.Context, int, model.TransactionCursor) ([]T, error)) {
	id, httpErr := extractUserId(r)
	if httpErr != nil {
		http.Error(w, httpErr.msg, httpErr.code)
		return
	}
	cursor, ok := transactionCursor(r)
	if !ok {
		http.Error(w, "invalid transaction cursor", http.StatusBadRequest)
		return
	}
	items, err := get(r.Context(), id, cursor)
	if err != nil {
		allInfoError(w, err)
		return
	}
	if items == nil {
		items = []T{}
	}
	w.Header().Set("Content-Type", "application/json")
	writeAndMarshal(w, items)
}

// @Summary Get deposit transactions
// @Description Returns up to 100 transactions, newest first. Supply before_id with before to include older IDs at the same timestamp; without before_id the timestamp is exclusive.
// @Tags transactions
// @Produce json
// @Param before query string false "Exclusive RFC3339 timestamp, or timestamp of cursor when before_id is set"
// @Param before_id query int false "Last transaction ID; requires before"
// @Success 200 {array} model.DepositTransaction
// @Failure 400 {string} string
// @Failure 401 {string} string
// @Failure 404 {string} string
// @Failure 500
// @Security BearerAuth
// @Router /getdeposittransactions [get]
func (h *handler) getDepositTransactions(w http.ResponseWriter, r *http.Request) {
	serveTransactionHistory(w, r, h.app.GetDepositTransactions)
}

// @Summary Get insurance transactions
// @Description Returns up to 100 transactions, newest first. Supply before_id with before to include older IDs at the same timestamp; without before_id the timestamp is exclusive.
// @Tags transactions
// @Produce json
// @Param before query string false "Exclusive RFC3339 timestamp, or timestamp of cursor when before_id is set"
// @Param before_id query int false "Last transaction ID; requires before"
// @Success 200 {array} model.InsuranceTransaction
// @Failure 400 {string} string
// @Failure 401 {string} string
// @Failure 404 {string} string
// @Failure 500
// @Security BearerAuth
// @Router /getinsurancetransactions [get]
func (h *handler) getInsuranceTransactions(w http.ResponseWriter, r *http.Request) {
	serveTransactionHistory(w, r, h.app.GetInsuranceTransactions)
}

// @Summary Get crypto transactions
// @Description Returns up to 100 transactions, newest first. Supply before_id with before to include older IDs at the same timestamp; without before_id the timestamp is exclusive.
// @Tags transactions
// @Produce json
// @Param before query string false "Exclusive RFC3339 timestamp, or timestamp of cursor when before_id is set"
// @Param before_id query int false "Last transaction ID; requires before"
// @Success 200 {array} model.CryptoTransaction
// @Failure 400 {string} string
// @Failure 401 {string} string
// @Failure 404 {string} string
// @Failure 500
// @Security BearerAuth
// @Router /getcryptotransactions [get]
func (h *handler) getCryptoTransactions(w http.ResponseWriter, r *http.Request) {
	serveTransactionHistory(w, r, h.app.GetCryptoTransactions)
}

// @Summary Get stock transactions
// @Description Returns up to 100 transactions, newest first. Supply before_id with before to include older IDs at the same timestamp; without before_id the timestamp is exclusive.
// @Tags transactions
// @Produce json
// @Param before query string false "Exclusive RFC3339 timestamp, or timestamp of cursor when before_id is set"
// @Param before_id query int false "Last transaction ID; requires before"
// @Success 200 {array} model.StockTransaction
// @Failure 400 {string} string
// @Failure 401 {string} string
// @Failure 404 {string} string
// @Failure 500
// @Security BearerAuth
// @Router /getstocktransactions [get]
func (h *handler) getStockTransactions(w http.ResponseWriter, r *http.Request) {
	serveTransactionHistory(w, r, h.app.GetStockTransactions)
}
