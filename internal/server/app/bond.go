package app

import (
	"context"
	"fmt"
	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/servererrors"
	"strings"
)

func (a *app) AddBonds(ctx context.Context, userId int, bonds []model.BondOperation) error {
	current, err := a.repo.GetBonds(ctx, userId)
	if err != nil {
		return err
	}
	var toInsert, toUpdate []model.BondOperation
	var amountDiff []int
	seen := make(map[string]bool, len(bonds))
	for _, bond := range bonds {
		if strings.TrimSpace(bond.Name) == "" {
			return servererrors.ErrInvalidBondName
		}
		if bond.LotAmount <= 0 {
			return fmt.Errorf("%w: %q", servererrors.ErrWrongBondAmount, bond.Code)
		}
		if bond.Price <= 0 {
			return fmt.Errorf("%w: %q", servererrors.ErrInvalidBondPrice, bond.Code)
		}
		if seen[bond.Code] {
			return fmt.Errorf("%w: %q", servererrors.ErrDuplicateBond, bond.Code)
		}
		seen[bond.Code] = true
		found := false
		for _, existing := range current {
			if bond.Code != existing.Code {
				continue
			}
			if bond.GroupName != existing.GroupName {
				return fmt.Errorf("%w: %q in %q", servererrors.ErrWrongBondGroup, bond.Code, existing.GroupName)
			}
			amountDiff = append(amountDiff, bond.LotAmount)
			bond.LotAmount += existing.LotAmount
			toUpdate = append(toUpdate, bond)
			found = true
			break
		}
		if !found {
			toInsert = append(toInsert, bond)
		}
	}
	if err := a.repo.InsertBonds(ctx, userId, toInsert); err != nil {
		return err
	}
	return a.repo.UpdateBondsAmount(ctx, userId, toUpdate, amountDiff)
}

func (a *app) RemoveBonds(ctx context.Context, userId int, bonds []model.BondOperation) error {
	current, err := a.repo.GetBonds(ctx, userId)
	if err != nil {
		return err
	}
	var toUpdate, toRemove []model.BondOperation
	var amountDiff []int
	seen := make(map[string]bool, len(bonds))
	for _, bond := range bonds {
		if strings.TrimSpace(bond.Name) == "" {
			return servererrors.ErrInvalidBondName
		}
		if bond.LotAmount <= 0 {
			return fmt.Errorf("%w: %q", servererrors.ErrWrongBondAmount, bond.Code)
		}
		if bond.Price <= 0 {
			return fmt.Errorf("%w: %q", servererrors.ErrInvalidBondPrice, bond.Code)
		}
		if seen[bond.Code] {
			return fmt.Errorf("%w: %q", servererrors.ErrDuplicateBond, bond.Code)
		}
		seen[bond.Code] = true
		found := false
		for _, existing := range current {
			if bond.Code != existing.Code {
				continue
			}
			if bond.GroupName != existing.GroupName {
				return fmt.Errorf("%w: %q in %q", servererrors.ErrWrongBondGroup, bond.Code, existing.GroupName)
			}
			if bond.LotAmount >= existing.LotAmount {
				bond.LotAmount = existing.LotAmount
				toRemove = append(toRemove, bond)
			} else {
				amountDiff = append(amountDiff, -bond.LotAmount)
				bond.LotAmount = existing.LotAmount - bond.LotAmount
				toUpdate = append(toUpdate, bond)
			}
			found = true
			break
		}
		if !found {
			return fmt.Errorf("%w: %q", servererrors.ErrBondNotFound, bond.Code)
		}
	}
	if err := a.repo.UpdateBondsAmount(ctx, userId, toUpdate, amountDiff); err != nil {
		return err
	}
	return a.repo.RemoveBonds(ctx, userId, toRemove)
}

func (a *app) GetBonds(ctx context.Context, userId int) ([]model.Bond, error) {
	bonds, err := a.repo.GetBonds(ctx, userId)
	if err != nil {
		return nil, err
	}
	codes := make([]string, len(bonds))
	for i := range bonds {
		codes[i] = bonds[i].Code
	}
	results := a.moex.GetBondsInfo(ctx, codes)
	if len(results) != len(bonds) {
		return nil, servererrors.ErrMoexUnhandled
	}
	for i, result := range results {
		if result.Err != nil {
			return nil, fmt.Errorf("%w: %w", servererrors.ErrMoexUnhandled, result.Err)
		}
		bonds[i].CleanPrice = result.CleanPrice
		bonds[i].AccruedInterest = result.AccruedInterest
		bonds[i].MaturityDate = result.MaturityDate
		bonds[i].MaturityPayment = result.MaturityPayment
		bonds[i].LotSize = result.LotSize
	}
	return bonds, nil
}

func (a *app) FindBonds(ctx context.Context, userId int, names []string) ([]model.FoundBond, error) {
	results := a.moex.FindBonds(ctx, names)
	if len(results) != len(names) {
		return nil, servererrors.ErrMoexUnhandled
	}
	bonds, err := a.repo.GetBonds(ctx, userId)
	if err != nil {
		return nil, err
	}
	found := make([]model.FoundBond, len(results))
	for i, result := range results {
		if result.Err != nil {
			return nil, fmt.Errorf("%w: %w", servererrors.ErrMoexUnhandled, result.Err)
		}
		if len(result.SearchResults) == 0 {
			found[i].ErrMsg = fmt.Sprintf("%s: %q", servererrors.ErrMoexBondNotFound, names[i])
			continue
		}
		for _, search := range result.SearchResults {
			if search.Err != nil {
				return nil, fmt.Errorf("%w: %w", servererrors.ErrMoexUnhandled, search.Err)
			}
			bond := search.Bond
			for _, existing := range bonds {
				if bond.Code == existing.Code {
					bond.GroupName = existing.GroupName
					bond.LotAmount = existing.LotAmount
					break
				}
			}
			found[i].SearchResults = append(found[i].SearchResults, bond)
		}
	}
	return found, nil
}
