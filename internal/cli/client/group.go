package client

import (
	"context"
	"net/http"
	"stocks_calculator/internal/model"
)

func (c *HttpClient) AddGroups(ctx context.Context, groups []model.Group) error {
	err := c.doJson(ctx, http.MethodPost, "/addgroups", groups, nil, true)
	return err
}

func (c *HttpClient) RemoveGroups(ctx context.Context, groups []model.Group) error {
	err := c.doJson(ctx, http.MethodDelete, "/removegroups", groups, nil, true)
	return err
}

func (c *HttpClient) RenameGroup(ctx context.Context, data model.RenameGroupBody) error {
	err := c.doJson(ctx, http.MethodPatch, "/renamegroup", data, nil, true)
	return err
}

func (c *HttpClient) GetGroups(ctx context.Context) ([]model.Group, error) {
	var result []model.Group
	err := c.doJson(ctx, http.MethodGet, "/getgroups", nil, &result, true)
	return result, err
}

func (c *HttpClient) UpdateWeights(ctx context.Context, groups []model.Group) error {
	err := c.doJson(ctx, http.MethodPatch, "/updateweights", groups, nil, true)
	return err
}
