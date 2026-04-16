package repo

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/servererrors"
	"time"
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
			`INSERT INTO stocks (group_id, name, code, amount)
			 SELECT id, $3, $4, $5
			 FROM groups
			 WHERE name = $1 AND profile_id = $2`,
			stock.GroupName, id, stock.Name, stock.Code, stock.Amount,
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

		err = registerStockTransaction(ctx, tx, model.Transaction{
			ProfileId: id,
			StockCode: stock.Code,
			Amount:    stock.Amount,
			Buying:    true,
			Price:     stock.Price,
			Time:      time.Now(),
		})
		if err != nil {
			return err
		}
	}

	tx.Commit()
	return nil
}

func (r *repo) UpdateStocksAmount(ctx context.Context, userId int, stocks []model.Stock, amountDiff []int) error {
	id, err := r.getDefaultProfileId(ctx, userId)
	if err != nil {
		return err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for i, stock := range stocks {
		result, err := tx.ExecContext(
			ctx,
			`UPDATE stocks SET amount = $1
			 WHERE code = $2
			 AND group_id = (SELECT id FROM groups WHERE name = $3 AND profile_id = $4)`,
			stock.Amount, stock.Code, stock.GroupName, id,
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

		err = registerStockTransaction(ctx, tx, model.Transaction{
			ProfileId: id,
			StockCode: stock.Code,
			Amount:    int(math.Abs(float64(amountDiff[i]))),
			Buying:    amountDiff[i] > 0,
			Price:     stock.Price,
			Time:      time.Now(),
		})
		if err != nil {
			return err
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

		err = registerStockTransaction(ctx, tx, model.Transaction{
			ProfileId: id,
			StockCode: stock.Code,
			Amount:    stock.Amount,
			Buying:    false,
			Price:     stock.Price,
			Time:      time.Now(),
		})
		if err != nil {
			return err
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
		`SELECT s.name, s.code, g.name, 0, s.amount
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

func registerStockTransaction(ctx context.Context, tx *sql.Tx, stTx model.Transaction) error {
	_, err := tx.ExecContext(
		ctx,
		`INSERT INTO transactions 
		 (profile_id, stock_code, amount, buying, price_in_rub, datetime)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		stTx.ProfileId, stTx.StockCode, stTx.Amount, stTx.Buying, stTx.Price, stTx.Time,
	)
	return err
}
