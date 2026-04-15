package app

import (
	"context"
	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/servererrors"
)

func (a *app) GetGroups(ctx context.Context, userId int) ([]model.Group, error) {
	return a.repo.GetGroups(ctx, userId)
}

func (a *app) AddGroups(ctx context.Context, userId int, groups []model.Group) error {
	for i := range groups {
		if len(groups[i].Name) > 64 {
			return servererrors.ErrInvalidGroupName
		}
	}
	return a.repo.InsertGroups(ctx, userId, groups)
}

func (a *app) RemoveGroups(ctx context.Context, userId int, groups []model.Group) error {
	return a.repo.RemoveGroups(ctx, userId, groups)
}

func (a *app) RenameGroup(ctx context.Context, userId int, group model.RenameGroupBody) error {
	return a.repo.UpdateGroupName(ctx, userId, group.NameOld, group.NameNew)
}

func (a *app) UpdateWeights(ctx context.Context, userId int, groups []model.Group) error {
	totalWeight := 0
	for i := range groups {
		if groups[i].Weight < 0 || groups[i].Weight > 100 {
			return servererrors.ErrInvalidWeights
		}
		totalWeight += groups[i].Weight
	}

	if totalWeight > 100 {
		return servererrors.ErrInvalidWeights
	}

	return a.repo.UpdateWeights(ctx, userId, groups)
}
