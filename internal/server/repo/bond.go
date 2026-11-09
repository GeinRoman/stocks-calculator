package repo

import (
	"context"
	"fmt"
	"math"
	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/servererrors"
	"time"
)

func (r *repo) InsertBonds(ctx context.Context, userId int, bonds []model.Bond) error {
	id, err := r.getDefaultProfileId(ctx, userId)
	if err != nil {
		return err
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, bond := range bonds {
		result, err := tx.ExecContext(ctx, `INSERT INTO bonds (group_id, name, code, lot_amount, lot_size)
			SELECT id, $3, $4, $5, $6 FROM bond_groups WHERE name = $1 AND profile_id = $2`,
			bond.GroupName, id, bond.Name, bond.Code, bond.LotAmount, bond.LotSize)
		if err != nil {
			return err
		}
		rows, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if rows == 0 {
			return fmt.Errorf("%w: %q", servererrors.ErrGroupNotFound, bond.GroupName)
		}
		if err := registerTransaction(ctx, tx, model.Transaction{ProfileId: id, InstrumentCode: bond.Code, InstrumentType: model.InstrumentBond, LotAmount: bond.LotAmount, Buying: true, LotSize: bond.LotSize, Price: bond.CleanPrice, Time: time.Now()}); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *repo) UpdateBondsAmount(ctx context.Context, userId int, bonds []model.Bond, amountDiff []int) error {
	if len(bonds) != len(amountDiff) {
		return fmt.Errorf("bonds and amount differences must have equal length")
	}
	id, err := r.getDefaultProfileId(ctx, userId)
	if err != nil {
		return err
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for i, bond := range bonds {
		result, err := tx.ExecContext(ctx, `UPDATE bonds SET lot_amount = $1 WHERE code = $2
			AND group_id = (SELECT id FROM bond_groups WHERE name = $3 AND profile_id = $4)`, bond.LotAmount, bond.Code, bond.GroupName, id)
		if err != nil {
			return err
		}
		rows, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if rows == 0 {
			return fmt.Errorf("failed to update bonds")
		}
		if err := registerTransaction(ctx, tx, model.Transaction{ProfileId: id, InstrumentCode: bond.Code, InstrumentType: model.InstrumentBond, LotAmount: int(math.Abs(float64(amountDiff[i]))), Buying: amountDiff[i] > 0, LotSize: bond.LotSize, Price: bond.CleanPrice, Time: time.Now()}); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *repo) RemoveBonds(ctx context.Context, userId int, bonds []model.Bond) error {
	id, err := r.getDefaultProfileId(ctx, userId)
	if err != nil {
		return err
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, bond := range bonds {
		result, err := tx.ExecContext(ctx, `DELETE FROM bonds WHERE code = $1
			AND group_id = (SELECT id FROM bond_groups WHERE name = $2 AND profile_id = $3)`, bond.Code, bond.GroupName, id)
		if err != nil {
			return err
		}
		rows, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if rows == 0 {
			return fmt.Errorf("failed to remove bonds")
		}
		if err := registerTransaction(ctx, tx, model.Transaction{ProfileId: id, InstrumentCode: bond.Code, InstrumentType: model.InstrumentBond, LotAmount: bond.LotAmount, Buying: false, LotSize: bond.LotSize, Price: bond.CleanPrice, Time: time.Now()}); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *repo) GetBonds(ctx context.Context, userId int) ([]model.Bond, error) {
	id, err := r.getDefaultProfileId(ctx, userId)
	if err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(ctx, `SELECT b.name, b.code, g.name, b.lot_amount, b.lot_size
		FROM bonds b JOIN bond_groups g ON b.group_id = g.id WHERE g.profile_id = $1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	bonds := []model.Bond{}
	for rows.Next() {
		var bond model.Bond
		if err := rows.Scan(&bond.Name, &bond.Code, &bond.GroupName, &bond.LotAmount, &bond.LotSize); err != nil {
			return nil, err
		}
		bonds = append(bonds, bond)
	}
	return bonds, rows.Err()
}
