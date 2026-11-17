package repo

import (
	"context"
	"fmt"
	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/servererrors"
)

func (r *repo) GetBondGroups(ctx context.Context, userId int) ([]model.BondGroup, error) {
	id, err := r.getDefaultProfileId(ctx, userId)
	if err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(ctx, `SELECT name, weight FROM bond_groups WHERE profile_id = $1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	groups := []model.BondGroup{}
	for rows.Next() {
		var group model.BondGroup
		if err := rows.Scan(&group.Name, &group.Weight); err != nil {
			return nil, err
		}
		groups = append(groups, group)
	}
	return groups, rows.Err()
}

func (r *repo) InsertBondGroups(ctx context.Context, userId int, groups []model.BondGroup) error {
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
		result, err := tx.ExecContext(ctx, `INSERT INTO bond_groups (profile_id, name) VALUES ($1, $2) ON CONFLICT (profile_id, name) DO NOTHING`, id, group.Name)
		if err != nil {
			return err
		}
		rows, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if rows == 0 {
			return fmt.Errorf("%w: %q", servererrors.ErrGroupExists, group.Name)
		}
	}
	return tx.Commit()
}

func (r *repo) RemoveBondGroups(ctx context.Context, userId int, groups []model.BondGroup) error {
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
		result, err := tx.ExecContext(ctx, `DELETE FROM bond_groups WHERE profile_id = $1 AND name = $2`, id, group.Name)
		if err != nil {
			return err
		}
		rows, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if rows == 0 {
			return fmt.Errorf("%w: %q", servererrors.ErrGroupNotFound, group.Name)
		}
	}
	return tx.Commit()
}

func (r *repo) UpdateBondGroupName(ctx context.Context, userId int, nameOld, nameNew string) error {
	id, err := r.getDefaultProfileId(ctx, userId)
	if err != nil {
		return err
	}
	result, err := r.db.ExecContext(ctx, `UPDATE bond_groups SET name = $2 WHERE profile_id = $3 AND name = $1`, nameOld, nameNew, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return servererrors.ErrGroupNotFound
	}
	return nil
}

func (r *repo) UpdateBondWeights(ctx context.Context, userId int, groups []model.BondGroup) error {
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
		result, err := tx.ExecContext(ctx, `UPDATE bond_groups SET weight = $1 WHERE profile_id = $2 AND name = $3`, group.Weight, id, group.Name)
		if err != nil {
			return err
		}
		rows, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if rows == 0 {
			return fmt.Errorf("%w: %q", servererrors.ErrGroupNotFound, group.Name)
		}
	}
	return tx.Commit()
}
