package app

import (
	"context"
	"math"
	"strings"

	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/servererrors"
)

func (s *app) GetCryptoGroups(ctx context.Context, user int) ([]model.CryptoGroup, error) {
	return s.repo.GetCryptoGroups(ctx, user)
}
func (s *app) AddCryptoGroups(ctx context.Context, user int, groups []model.CryptoGroup) error {
	for _, g := range groups {
		if strings.TrimSpace(g.Name) == "" || len(g.Name) > 64 {
			return servererrors.ErrInvalidGroupName
		}
	}
	return s.repo.InsertCryptoGroups(ctx, user, groups)
}
func (s *app) RemoveCryptoGroups(ctx context.Context, user int, groups []model.CryptoGroup) error {
	return s.repo.RemoveCryptoGroups(ctx, user, groups)
}
func (s *app) RenameCryptoGroup(ctx context.Context, user int, group model.RenameCryptoGroupBody) error {
	if strings.TrimSpace(group.NameOld) == "" || strings.TrimSpace(group.NameNew) == "" || len(group.NameNew) > 64 {
		return servererrors.ErrInvalidGroupName
	}
	return s.repo.UpdateCryptoGroupName(ctx, user, group.NameOld, group.NameNew)
}
func (s *app) UpdateCryptoWeights(ctx context.Context, user int, groups []model.CryptoGroup) error {
	total := 0
	seen := map[string]bool{}
	for _, g := range groups {
		if g.Weight < 0 || g.Weight > 100 || seen[g.Name] {
			return servererrors.ErrInvalidWeights
		}
		seen[g.Name] = true
		total += g.Weight
	}
	if total > 100 {
		return servererrors.ErrInvalidWeights
	}
	return s.repo.UpdateCryptoWeights(ctx, user, groups)
}
func (s *app) GetCryptoAssets(ctx context.Context, user int) ([]model.CryptoAsset, error) {
	return s.repo.GetCryptoAssets(ctx, user)
}
func (s *app) UpdateCryptoAsset(ctx context.Context, user int, asset model.CryptoAssetUpdate) error {
	if asset.ID <= 0 || !validCryptoMetadata(asset.Name, asset.Platform, asset.AnnualInterest) {
		return servererrors.ErrInvalidCrypto
	}
	return s.repo.UpdateCryptoAsset(ctx, user, asset)
}

func validCryptoMetadata(name, platform string, interest float64) bool {
	return strings.TrimSpace(name) != "" && strings.TrimSpace(platform) != "" && interest >= 0 && !math.IsNaN(interest) && !math.IsInf(interest, 0)
}

func validCryptoPrice(price model.Money) bool {
	return price.Price > 0 && !math.IsNaN(price.Price) && !math.IsInf(price.Price, 0) && strings.TrimSpace(price.Currency) != "" && len(price.Currency) <= 16
}
func (s *app) BuyCrypto(ctx context.Context, user int, ops []model.CryptoOperation) error {
	return s.transact(ctx, user, ops, true)
}
func (s *app) SellCrypto(ctx context.Context, user int, ops []model.CryptoOperation) error {
	return s.transact(ctx, user, ops, false)
}
func (s *app) transact(ctx context.Context, user int, ops []model.CryptoOperation, buy bool) error {
	seenIDs := map[int64]bool{}
	seenNames := map[string]bool{}
	for _, op := range ops {
		if op.ID < 0 || (!buy && op.ID == 0) || op.Amount <= 0 || math.IsNaN(op.Amount) || math.IsInf(op.Amount, 0) || !validCryptoPrice(op.Price) {
			return servererrors.ErrInvalidCrypto
		}
		if op.ID == 0 {
			if !validCryptoMetadata(op.Name, op.Platform, op.AnnualInterest) || strings.TrimSpace(op.GroupName) == "" || seenNames[op.Name] {
				return servererrors.ErrInvalidCrypto
			}
			seenNames[op.Name] = true
		} else {
			if seenIDs[op.ID] {
				return servererrors.ErrInvalidCrypto
			}
			seenIDs[op.ID] = true
		}
	}
	return s.repo.TransactCrypto(ctx, user, ops, buy)
}
