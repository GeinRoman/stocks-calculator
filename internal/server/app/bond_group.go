package app

import (
	"context"
	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/servererrors"
)

func (a *app) GetBondGroups(ctx context.Context, userId int) ([]model.BondGroup, error) {
	return a.repo.GetBondGroups(ctx, userId)
}

func (a *app) AddBondGroups(ctx context.Context, userId int, groups []model.BondGroup) error {
	for _, group := range groups {
		if len(group.Name) > 64 {
			return servererrors.ErrInvalidGroupName
		}
	}
	return a.repo.InsertBondGroups(ctx, userId, groups)
}

func (a *app) RemoveBondGroups(ctx context.Context, userId int, groups []model.BondGroup) error {
	return a.repo.RemoveBondGroups(ctx, userId, groups)
}

func (a *app) RenameBondGroup(ctx context.Context, userId int, group model.RenameBondGroupBody) error {
	return a.repo.UpdateBondGroupName(ctx, userId, group.NameOld, group.NameNew)
}

func (a *app) UpdateBondWeights(ctx context.Context, userId int, groups []model.BondGroup) error {
	total := 0
	for _, group := range groups {
		if group.Weight < 0 || group.Weight > 100 {
			return servererrors.ErrInvalidWeights
		}
		total += group.Weight
	}
	if total > 100 {
		return servererrors.ErrInvalidWeights
	}
	return a.repo.UpdateBondWeights(ctx, userId, groups)
}
