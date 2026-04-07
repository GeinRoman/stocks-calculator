package repo

import (
	"context"
	"stocks_calculator/internal/model"
)

func (r *Repo) Profiles(ctx context.Context, userId int) ([]model.Profile, error) {
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
