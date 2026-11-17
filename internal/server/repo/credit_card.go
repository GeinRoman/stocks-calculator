package repo

import (
	"context"
	"database/sql"

	"stocks_calculator/internal/model"
)

func (r *repo) GetCreditCards(ctx context.Context, user int) ([]model.CreditCard, error) {
	profile, err := r.personalAssetProfile(ctx, user)
	if err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, bank_name, to_char(mandatory_payment_date, 'YYYY-MM-DD'), mandatory_payment_amount_rub, overall_limit_rub FROM credit_cards WHERE profile_id=$1 ORDER BY id`, profile)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cards := []model.CreditCard{}
	for rows.Next() {
		var card model.CreditCard
		if err := rows.Scan(&card.ID, &card.Name, &card.BankName, &card.MandatoryPaymentDate, &card.MandatoryPaymentAmountRUB, &card.OverallLimitRUB); err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}
	return cards, rows.Err()
}

func (r *repo) AddCreditCard(ctx context.Context, user int, card model.CreditCard) error {
	return r.personalAssetWrite(ctx, user, func(tx *sql.Tx, profile int) error {
		_, err := tx.ExecContext(ctx, `INSERT INTO credit_cards (profile_id, name, bank_name, mandatory_payment_date, mandatory_payment_amount_rub, overall_limit_rub) VALUES ($1,$2,$3,$4,$5,$6)`, profile, card.Name, card.BankName, card.MandatoryPaymentDate, card.MandatoryPaymentAmountRUB, card.OverallLimitRUB)
		return err
	})
}

func (r *repo) UpdateCreditCard(ctx context.Context, user int, card model.CreditCardUpdate) error {
	return r.personalAssetWrite(ctx, user, func(tx *sql.Tx, profile int) error {
		result, err := tx.ExecContext(ctx, `UPDATE credit_cards SET name=COALESCE($1,name), bank_name=COALESCE($2,bank_name), mandatory_payment_date=COALESCE($3::date,mandatory_payment_date), mandatory_payment_amount_rub=COALESCE($4,mandatory_payment_amount_rub), overall_limit_rub=COALESCE($5,overall_limit_rub) WHERE profile_id=$6 AND id=$7`, card.Name, card.BankName, card.MandatoryPaymentDate, card.MandatoryPaymentAmountRUB, card.OverallLimitRUB, profile, card.ID)
		return creditCardMutationResult(result, err)
	})
}

func (r *repo) RemoveCreditCard(ctx context.Context, user int, id int64) error {
	return r.personalAssetWrite(ctx, user, func(tx *sql.Tx, profile int) error {
		result, err := tx.ExecContext(ctx, `DELETE FROM credit_cards WHERE profile_id=$1 AND id=$2`, profile, id)
		return creditCardMutationResult(result, err)
	})
}

func creditCardMutationResult(result sql.Result, err error) error {
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return model.ErrCreditCardNotFound
	}
	return nil
}
