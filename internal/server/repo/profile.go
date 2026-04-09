package repo

import (
	"context"
	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/app"
)

func (r *repo) GetProfiles(ctx context.Context, userId int) ([]model.Profile, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT name, is_default FROM profiles WHERE user_id = $1",
		userId,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	profiles, err := getTable[model.Profile](rows)
	if err != nil {
		return nil, err
	}
	return profiles, nil
}

func (r *repo) RemoveProfile(ctx context.Context, profile string, userId int) error {
	result, err := r.db.ExecContext(ctx,
		"DELETE FROM profiles WHERE user_id = $1 AND name = $2",
		userId, profile,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return app.ErrProfileNotFound
	}

	return nil
}

func (r *repo) InsertProfile(ctx context.Context, profile string, userId int, def bool) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx,
		`INSERT INTO profiles (user_id, name, is_default)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (user_id, name) DO NOTHING`,
		userId, profile, def,
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return app.ErrProfileExists
	}

	if def {
		_, err = tx.ExecContext(ctx,
			`UPDATE profiles
		 SET is_default = (name = $1)
		 WHERE user_id = $2`,
			profile, userId,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *repo) SetDefaultProfile(ctx context.Context, profile string, userId int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE profiles
		 SET is_default = (name = $1)
		 WHERE user_id = $2`,
		profile, userId,
	)
	return err
}
