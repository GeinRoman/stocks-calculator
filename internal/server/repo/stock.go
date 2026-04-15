package repo

import (
	"context"
	"fmt"
	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/servererrors"
)

func (r *repo) InsertStocks(ctx context.Context, userId int, stocks []model.Stock) error {
	id, err := r.getDefaultProfileId(ctx, userId)
	if err != nil {
		return err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, stock := range stocks {
		result, err := tx.ExecContext(
			ctx,
			`INSERT INTO stocks (group_id, name, code, lot_amount, lot_size)
			 SELECT id, $3, $4, $5, $6
			 FROM groups
			 WHERE name = $1 AND profile_id = $2`,
			stock.GroupName, id, stock.Name, stock.Code, stock.LotAmount, stock.LotSize,
		)
		if err != nil {
			return err
		}
		rows, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if rows == 0 {
			return fmt.Errorf("%w: %q", servererrors.ErrGroupNotFound, stock.GroupName)
		}
	}

	tx.Commit()
	return nil
}

func (r *repo) UpdateStocksAmount(ctx context.Context, userId int, stocks []model.Stock) error {
	id, err := r.getDefaultProfileId(ctx, userId)
	if err != nil {
		return err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, stock := range stocks {
		result, err := tx.ExecContext(
			ctx,
			`UPDATE stocks SET lot_amount = $1
			 WHERE code = $2
			 AND group_id = (SELECT id FROM groups WHERE name = $3 AND profile_id = $4)`,
			stock.LotAmount, stock.Code, stock.GroupName, id,
		)
		if err != nil {
			return err
		}
		rows, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if rows == 0 {
			return fmt.Errorf("Failed to update stocks")
		}
	}

	tx.Commit()
	return nil
}

func (r *repo) RemoveStocks(ctx context.Context, userId int, stocks []model.Stock) error {
	id, err := r.getDefaultProfileId(ctx, userId)
	if err != nil {
		return err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, stock := range stocks {
		result, err := tx.ExecContext(
			ctx,
			`DELETE FROM stocks
			 WHERE code = $1
			 AND group_id = (SELECT id FROM groups WHERE name = $2 AND profile_id = $3)`,
			stock.Code, stock.GroupName, id,
		)
		if err != nil {
			return err
		}
		rows, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if rows == 0 {
			return fmt.Errorf("Failed to remove stocks")
		}
	}

	tx.Commit()
	return nil
}

func (r *repo) GetStocks(ctx context.Context, userId int) ([]model.Stock, error) {
	id, err := r.getDefaultProfileId(ctx, userId)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(
		ctx,
		`SELECT s.name, s.code, g.name, 0, s.lot_amount, s.lot_size
		 FROM stocks s JOIN groups g ON s.group_id = g.id
		 WHERE g.profile_id = $1`,
		id,
	)
	if err != nil {
		return nil, err
	}
	stocks, err := getTable[model.Stock](rows)
	if err != nil {
		return nil, err
	}

	return stocks, nil
}
