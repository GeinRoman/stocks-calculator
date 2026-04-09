package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/app"
)

func (r *repo) GetGroups(ctx context.Context, userId int) ([]model.Group, error) {
	id, err := r.getDefaultProfileId(ctx, userId)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(
		ctx,
		"SELECT name, weight FROM groups WHERE profile_id = $1",
		id,
	)
	if err != nil {
		return nil, err
	}
	profiles, err := getTable[model.Group](rows)
	if err != nil {
		return nil, err
	}

	return profiles, nil
}

func (r *repo) InsertGroups(ctx context.Context, userId int, groups []model.Group) error {
	id, err := r.getDefaultProfileId(ctx, userId)
	if err != nil {
		return err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, group := range groups {
		result, err := tx.ExecContext(
			ctx,
			`INSERT INTO groups (profile_id, name) VALUES ($1, $2)
			ON CONFLICT (profile_id, name) DO NOTHING`,
			id, group.Name,
		)
		if err != nil {
			return err
		}
		rows, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if rows == 0 {
			return fmt.Errorf("%w: %q", app.ErrGroupExists, group.Name)
		}
	}

	tx.Commit()
	return nil
}

func (r *repo) UpdateGroupName(ctx context.Context, userId int, nameOld, nameNew string) error {
	id, err := r.getDefaultProfileId(ctx, userId)
	if err != nil {
		return err
	}

	result, err := r.db.ExecContext(
		ctx,
		`UPDATE groups SET name = $2 WHERE profile_id = $3 AND name = $1`,
		nameOld, nameNew, id,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return app.ErrGroupNotFound
	}

	return nil
}

func (r *repo) RemoveGroups(ctx context.Context, userId int, groups []model.Group) error {
	id, err := r.getDefaultProfileId(ctx, userId)
	if err != nil {
		return err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, group := range groups {
		result, err := tx.ExecContext(
			ctx,
			`DELETE FROM groups WHERE profile_id = $1 AND name = $2`,
			id, group.Name,
		)
		if err != nil {
			return err
		}
		rows, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if rows == 0 {
			return fmt.Errorf("%w: %q", app.ErrGroupNotFound, group.Name)
		}
	}

	tx.Commit()
	return nil
}

func (r *repo) UpdateWeights(ctx context.Context, userId int, groups []model.Group) error {
	id, err := r.getDefaultProfileId(ctx, userId)
	if err != nil {
		return err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, group := range groups {
		result, err := tx.ExecContext(
			ctx,
			"UPDATE groups SET weight = $1 WHERE profile_id = $2 AND name = $3",
			group.Weight, id, group.Name,
		)
		if err != nil {
			return err
		}
		rows, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if rows == 0 {
			return fmt.Errorf("%w: %q", app.ErrGroupNotFound, group.Name)
		}
	}

	tx.Commit()
	return err
}

func (r *repo) getDefaultProfileId(ctx context.Context, userId int) (int, error) {
	var profileId int
	err := r.db.QueryRowContext(
		ctx,
		"SELECT id FROM profiles WHERE user_id = $1 AND is_default = $2",
		userId, true,
	).Scan(&profileId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, app.ErrNoDefaultProfile
		}
		return 0, err
	}
	return profileId, nil
}
