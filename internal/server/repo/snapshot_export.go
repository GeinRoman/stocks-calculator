package repo

import (
	"context"
	"database/sql"
	"encoding/json"

	"stocks_calculator/internal/model"
)

func snapshotList(ctx context.Context, tx *sql.Tx, profile int, query string, dest any) error {
	var data []byte
	if err := tx.QueryRowContext(ctx, query, profile).Scan(&data); err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func (r *repo) GetSnapshot(ctx context.Context, user int) (model.Snapshot, error) {
	result := model.Snapshot{Version: 1, Profiles: []model.SnapshotProfile{}}
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return result, err
	}
	defer tx.Rollback()
	rows, err := tx.QueryContext(ctx, `SELECT id,name,is_default FROM profiles WHERE user_id=$1 ORDER BY id`, user)
	if err != nil {
		return result, err
	}
	var ids []int
	for rows.Next() {
		var id int
		var p model.SnapshotProfile
		if err = rows.Scan(&id, &p.Name, &p.Default); err != nil {
			rows.Close()
			return result, err
		}
		ids = append(ids, id)
		result.Profiles = append(result.Profiles, p)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return result, err
	}
	for i, id := range ids {
		p := &result.Profiles[i]
		queries := []struct {
			query  string
			target any
		}{
			{`SELECT COALESCE(json_agg(row_to_json(x) ORDER BY x.id),'[]'::json) FROM (SELECT g.id,g.name,g.weight, COALESCE((SELECT json_agg(json_build_object('name',s.name,'code',s.code,'lot_amount',s.lot_amount,'lot_size',s.lot_size) ORDER BY s.id) FROM stocks s WHERE s.group_id=g.id),'[]'::json) AS stocks FROM groups g WHERE g.profile_id=$1) x`, &p.StockGroups},
			{`SELECT COALESCE(json_agg(row_to_json(x) ORDER BY x.id),'[]'::json) FROM (SELECT g.id,g.name,g.weight, COALESCE((SELECT json_agg(json_build_object('name',b.name,'code',b.code,'lot_amount',b.lot_amount,'lot_size',b.lot_size) ORDER BY b.id) FROM bonds b WHERE b.group_id=g.id),'[]'::json) AS bonds FROM bond_groups g WHERE g.profile_id=$1) x`, &p.BondGroups},
			{`SELECT COALESCE(json_agg(row_to_json(x) ORDER BY x.id),'[]'::json) FROM (SELECT g.id,g.name,g.weight, COALESCE((SELECT json_agg(json_build_object('id',a.id,'name',a.name,'platform',a.platform,'url',a.url,'annual_interest',a.annual_interest,'amount',a.amount::text) ORDER BY a.id) FROM crypto_assets a WHERE a.group_id=g.id AND a.profile_id=g.profile_id),'[]'::json) AS assets FROM crypto_groups g WHERE g.profile_id=$1) x`, &p.CryptoGroups},
			{`SELECT COALESCE(json_agg(row_to_json(x) ORDER BY x.id),'[]'::json) FROM (SELECT id,name,bank_name,amount_rub,annual_interest,start_date::text,end_date::text,acquisition_transaction_id FROM deposits WHERE profile_id=$1) x`, &p.Deposits},
			{`SELECT COALESCE(json_agg(row_to_json(x) ORDER BY x.id),'[]'::json) FROM (SELECT id,name,company_name,amount_rub,annual_interest,start_date::text,end_date::text,acquisition_transaction_id FROM insurances WHERE profile_id=$1) x`, &p.Insurances},
			{`SELECT COALESCE(json_agg(row_to_json(x) ORDER BY x.id),'[]'::json) FROM (SELECT id,name,bank_name,mandatory_payment_date::text,mandatory_payment_amount_rub,overall_limit_rub FROM credit_cards WHERE profile_id=$1) x`, &p.CreditCards},
			{`SELECT COALESCE(json_agg(row_to_json(x) ORDER BY x.id),'[]'::json) FROM (SELECT id,instrument_code,instrument_type,lot_amount,buying,lot_size,price,currency,datetime AS time FROM transactions WHERE profile_id=$1) x`, &p.Transactions},
			{`SELECT COALESCE(json_agg(row_to_json(x) ORDER BY x.id),'[]'::json) FROM (SELECT id,asset_id,asset_name,platform,amount::text,buying,price,currency,created_at FROM crypto_transactions WHERE profile_id=$1) x`, &p.CryptoTransactions},
			{`SELECT COALESCE(json_agg(row_to_json(x) ORDER BY x.id),'[]'::json) FROM (SELECT id,name,bank_name,amount_rub,buying,created_at FROM deposit_transactions WHERE profile_id=$1) x`, &p.DepositTransactions},
			{`SELECT COALESCE(json_agg(row_to_json(x) ORDER BY x.id),'[]'::json) FROM (SELECT id,name,company_name,amount_rub,buying,created_at FROM insurance_transactions WHERE profile_id=$1) x`, &p.InsuranceTransactions},
		}
		for _, q := range queries {
			if err := snapshotList(ctx, tx, id, q.query, q.target); err != nil {
				return result, err
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return result, err
	}
	return result, nil
}
