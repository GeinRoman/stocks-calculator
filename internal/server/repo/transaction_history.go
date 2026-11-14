package repo

import (
	"context"
	"database/sql"

	"stocks_calculator/internal/model"
)

// The optional ID makes a timestamp boundary inclusive only for smaller IDs.
func readTransactionHistory[T any](ctx context.Context, r *repo, user int, cursor model.TransactionCursor, query string, scan func(*sql.Rows, *T) error) ([]T, error) {
	profile, err := r.getDefaultProfileId(ctx, user)
	if err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(ctx, query, profile, cursor.Before, cursor.BeforeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []T{}
	for rows.Next() {
		var item T
		if err := scan(rows, &item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *repo) GetDepositTransactions(ctx context.Context, user int, cursor model.TransactionCursor) ([]model.DepositTransaction, error) {
	return readTransactionHistory(ctx, r, user, cursor, `SELECT id,name,bank_name,amount_rub,buying,created_at FROM deposit_transactions
		WHERE profile_id=$1 AND ($2::timestamptz IS NULL OR created_at < $2 OR ($3::bigint IS NOT NULL AND created_at = $2 AND id < $3))
		ORDER BY created_at DESC,id DESC LIMIT 100`, func(rows *sql.Rows, item *model.DepositTransaction) error {
		return rows.Scan(&item.ID, &item.Name, &item.BankName, &item.AmountRUB, &item.Buying, &item.CreatedAt)
	})
}
func (r *repo) GetInsuranceTransactions(ctx context.Context, user int, cursor model.TransactionCursor) ([]model.InsuranceTransaction, error) {
	return readTransactionHistory(ctx, r, user, cursor, `SELECT id,name,company_name,amount_rub,buying,created_at FROM insurance_transactions
		WHERE profile_id=$1 AND ($2::timestamptz IS NULL OR created_at < $2 OR ($3::bigint IS NOT NULL AND created_at = $2 AND id < $3))
		ORDER BY created_at DESC,id DESC LIMIT 100`, func(rows *sql.Rows, item *model.InsuranceTransaction) error {
		return rows.Scan(&item.ID, &item.Name, &item.CompanyName, &item.AmountRUB, &item.Buying, &item.CreatedAt)
	})
}
func (r *repo) GetCryptoTransactions(ctx context.Context, user int, cursor model.TransactionCursor) ([]model.CryptoTransaction, error) {
	return readTransactionHistory(ctx, r, user, cursor, `SELECT id,asset_id,asset_name,platform,amount,buying,price,currency,created_at FROM crypto_transactions
		WHERE profile_id=$1 AND ($2::timestamptz IS NULL OR created_at < $2 OR ($3::bigint IS NOT NULL AND created_at = $2 AND id < $3))
		ORDER BY created_at DESC,id DESC LIMIT 100`, func(rows *sql.Rows, item *model.CryptoTransaction) error {
		return rows.Scan(&item.ID, &item.AssetID, &item.AssetName, &item.Platform, &item.Amount, &item.Buying, &item.Price, &item.Currency, &item.CreatedAt)
	})
}
func (r *repo) GetStockTransactions(ctx context.Context, user int, cursor model.TransactionCursor) ([]model.StockTransaction, error) {
	return readTransactionHistory(ctx, r, user, cursor, `SELECT id,instrument_code,lot_amount,buying,lot_size,price,currency,datetime FROM transactions
		WHERE profile_id=$1 AND instrument_type='stock' AND ($2::timestamptz IS NULL OR datetime < $2 OR ($3::bigint IS NOT NULL AND datetime = $2 AND id < $3))
		ORDER BY datetime DESC,id DESC LIMIT 100`, func(rows *sql.Rows, item *model.StockTransaction) error {
		return rows.Scan(&item.ID, &item.InstrumentCode, &item.LotAmount, &item.Buying, &item.LotSize, &item.Price, &item.Currency, &item.Time)
	})
}
