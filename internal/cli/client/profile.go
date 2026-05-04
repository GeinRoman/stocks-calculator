package client

import (
	"context"
	"net/http"
	"stocks_calculator/internal/model"
)

func (c *HttpClient) RemoveProfile(ctx context.Context, profile model.Profile) error {
	err := c.doJson(ctx, http.MethodDelete, "/removeprofile", profile, nil, true)
	return err
}

func (c *HttpClient) CreateProfile(ctx context.Context, profile model.Profile) error {
	err := c.doJson(
		ctx,
		http.MethodPost,
		"/createprofile",
		profile,
		nil,
		true,
	)
	return err
}

func (c *HttpClient) GetProfiles(ctx context.Context) ([]model.Profile, error) {
	var profiles []model.Profile
	err := c.doJson(ctx, http.MethodGet, "/getprofiles", nil, &profiles, true)
	return profiles, err
}

func (c *HttpClient) SetDefaultProfile(ctx context.Context, profile model.Profile) error {
	err := c.doJson(ctx, http.MethodPatch, "/setdefaultprofile", profile, nil, true)
	return err
}
