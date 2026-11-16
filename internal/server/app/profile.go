package app

import (
	"context"
	"stocks_calculator/internal/model"
)

func (a *app) RemoveProfile(ctx context.Context, profile model.Profile, userId int) error {
	err := a.repo.RemoveProfile(ctx, profile.Name, userId)
	if err != nil {
		return err
	}
	return nil
}

func (a *app) CreateProfile(ctx context.Context, profile model.Profile, userId int) error {
	err := a.repo.InsertProfile(ctx, profile.Name, userId, profile.Default)
	if err != nil {
		return err
	}
	return nil
}

func (a *app) GetProfiles(ctx context.Context, userId int) ([]model.Profile, error) {
	profiles, err := a.repo.GetProfiles(ctx, userId)
	if err != nil {
		return nil, err
	}
	return profiles, nil
}

func (a *app) SetDefaultProfile(ctx context.Context, profile model.Profile, userId int) error {
	return a.repo.SetDefaultProfile(ctx, profile.Name, userId)
}

func (a *app) RenameProfile(ctx context.Context, profile model.RenameProfileBody, userId int) error {
	return a.repo.RenameProfile(ctx, profile.Name, profile.NewName, userId)
}
