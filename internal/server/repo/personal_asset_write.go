package repo

import (
	"context"
	"database/sql"
	"errors"

	"stocks_calculator/internal/server/servererrors"
)

func (r *repo) personalAssetWrite(ctx context.Context, user int, fn func(*sql.Tx, int) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var profile int
	err = tx.QueryRowContext(ctx, `SELECT id FROM profiles WHERE user_id=$1 AND is_default=true FOR UPDATE`, user).Scan(&profile)
	if errors.Is(err, sql.ErrNoRows) {
		return servererrors.ErrNoDefaultProfile
	}
	if err != nil {
		return err
	}
	if err = fn(tx, profile); err != nil {
		return err
	}
	return tx.Commit()
}
