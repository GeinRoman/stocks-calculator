package app

import (
	"context"
	"stocks_calculator/internal/model"
)

func (a *app) GetDepositTransactions(ctx context.Context, user int, cursor model.TransactionCursor) ([]model.DepositTransaction, error) {
	return a.repo.GetDepositTransactions(ctx, user, cursor)
}
func (a *app) GetInsuranceTransactions(ctx context.Context, user int, cursor model.TransactionCursor) ([]model.InsuranceTransaction, error) {
	return a.repo.GetInsuranceTransactions(ctx, user, cursor)
}
func (a *app) GetCryptoTransactions(ctx context.Context, user int, cursor model.TransactionCursor) ([]model.CryptoTransaction, error) {
	return a.repo.GetCryptoTransactions(ctx, user, cursor)
}
func (a *app) GetStockTransactions(ctx context.Context, user int, cursor model.TransactionCursor) ([]model.StockTransaction, error) {
	return a.repo.GetStockTransactions(ctx, user, cursor)
}
