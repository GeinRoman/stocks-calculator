package repo

import (
	"context"
	"database/sql"

	"github.com/lib/pq"

	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/servererrors"
)

func contractDateError(err error) error {
	if e, ok := err.(*pq.Error); ok && e.Code == "23514" && (e.Constraint == "deposits_date_order" || e.Constraint == "insurances_date_order") {
		return servererrors.ErrInvalidContract
	}
	return err
}

func (r *repo) GetDeposits(ctx context.Context, user int) ([]model.Deposit, error) {
	profile, err := r.personalAssetProfile(ctx, user)
	if err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id,name,bank_name,amount_rub,annual_interest,start_date::text,end_date::text FROM deposits WHERE profile_id=$1 ORDER BY id`, profile)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []model.Deposit{}
	for rows.Next() {
		var d model.Deposit
		if err := rows.Scan(&d.ID, &d.Name, &d.BankName, &d.AmountRUB, &d.AnnualInterest, &d.StartDate, &d.EndDate); err != nil {
			return nil, err
		}
		items = append(items, d)
	}
	return items, rows.Err()
}

func (r *repo) AddDeposit(ctx context.Context, user int, d model.Deposit) error {
	return r.personalAssetWrite(ctx, user, func(tx *sql.Tx, profile int) error {
		var id int64
		if err := tx.QueryRowContext(ctx, `INSERT INTO deposit_transactions (profile_id,name,bank_name,amount_rub,buying,created_at) VALUES ($1,$2,$3,$4,true,CURRENT_TIMESTAMP) RETURNING id`, profile, d.Name, d.BankName, d.AmountRUB).Scan(&id); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx, `INSERT INTO deposits (profile_id,name,bank_name,amount_rub,annual_interest,start_date,end_date,acquisition_transaction_id) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, profile, d.Name, d.BankName, d.AmountRUB, d.AnnualInterest, d.StartDate, d.EndDate, id)
		return contractDateError(err)
	})
}

func (r *repo) UpdateDeposit(ctx context.Context, user int, d model.DepositUpdate) error {
	return r.personalAssetWrite(ctx, user, func(tx *sql.Tx, profile int) error {
		var id int64
		if err := tx.QueryRowContext(ctx, `UPDATE deposits SET name=COALESCE($1,name),bank_name=COALESCE($2,bank_name),annual_interest=COALESCE($3,annual_interest),end_date=COALESCE($4::date,end_date),start_date=COALESCE($5::date,start_date) WHERE profile_id=$6 AND id=$7 RETURNING acquisition_transaction_id`, d.Name, d.BankName, d.AnnualInterest, d.EndDate, d.StartDate, profile, d.ID).Scan(&id); err != nil {
			return contractDateError(err)
		}
		res, err := tx.ExecContext(ctx, `UPDATE deposit_transactions SET name=COALESCE($1,name),bank_name=COALESCE($2,bank_name) WHERE profile_id=$3 AND id=$4`, d.Name, d.BankName, profile, id)
		if err != nil {
			return err
		}
		n, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if n != 1 {
			return sql.ErrNoRows
		}
		return nil
	})
}

func (r *repo) RemoveDeposit(ctx context.Context, user int, id int64) error {
	return r.personalAssetWrite(ctx, user, func(tx *sql.Tx, profile int) error {
		var acquisition int64
		if err := tx.QueryRowContext(ctx, `DELETE FROM deposits WHERE profile_id=$1 AND id=$2 RETURNING acquisition_transaction_id`, profile, id).Scan(&acquisition); err != nil {
			return err
		}
		res, err := tx.ExecContext(ctx, `DELETE FROM deposit_transactions WHERE profile_id=$1 AND id=$2`, profile, acquisition)
		if err != nil {
			return err
		}
		n, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if n != 1 {
			return sql.ErrNoRows
		}
		return nil
	})
}

func (r *repo) CloseDeposit(ctx context.Context, user int, d model.DepositClose) error {
	return r.personalAssetWrite(ctx, user, func(tx *sql.Tx, profile int) error {
		var name, bank string
		if err := tx.QueryRowContext(ctx, `DELETE FROM deposits WHERE profile_id=$1 AND id=$2 RETURNING name,bank_name`, profile, d.ID).Scan(&name, &bank); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx, `INSERT INTO deposit_transactions (profile_id,name,bank_name,amount_rub,buying,created_at) VALUES ($1,$2,$3,$4,false,CURRENT_TIMESTAMP)`, profile, name, bank, d.ClosingAmountRUB)
		return err
	})
}
