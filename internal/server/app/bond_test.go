package app

import (
	"context"
	"errors"
	"testing"

	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/servererrors"
)

type bondTestRepo struct {
	Repository
	bonds                      []model.Bond
	writes                     int
	inserted, updated, removed []model.BondOperation
	diffs                      []int
}

func (r *bondTestRepo) GetBonds(context.Context, int) ([]model.Bond, error) {
	return append([]model.Bond(nil), r.bonds...), nil
}
func (r *bondTestRepo) InsertBonds(_ context.Context, _ int, bonds []model.BondOperation) error {
	r.writes++
	r.inserted = append(r.inserted, bonds...)
	return nil
}
func (r *bondTestRepo) UpdateBondsAmount(_ context.Context, _ int, bonds []model.BondOperation, diffs []int) error {
	r.writes++
	r.updated = append(r.updated, bonds...)
	r.diffs = append(r.diffs, diffs...)
	return nil
}
func (r *bondTestRepo) RemoveBonds(_ context.Context, _ int, bonds []model.BondOperation) error {
	r.writes++
	r.removed = append(r.removed, bonds...)
	return nil
}
func (r *bondTestRepo) UpdateBondWeights(context.Context, int, []model.BondGroup) error {
	r.writes++
	return nil
}

func TestBondOperationsPreserveTransactionPrice(t *testing.T) {
	repo := &bondTestRepo{bonds: []model.Bond{{Code: "A", GroupName: "G", LotAmount: 3}, {Code: "B", GroupName: "G", LotAmount: 2}}}
	// A nil MOEX dependency ensures mutations do not fetch market data.
	a := New(repo, nil, nil)
	add := []model.BondOperation{{Name: "A", Code: "A", GroupName: "G", LotAmount: 1, LotSize: 10, Price: 103.5}, {Name: "C", Code: "C", GroupName: "G", LotAmount: 2, LotSize: 1, Price: 98.75}}
	if err := a.AddBonds(context.Background(), 1, add); err != nil {
		t.Fatal(err)
	}
	if len(repo.updated) != 1 || repo.updated[0].Price != 103.5 || repo.updated[0].LotAmount != 4 || len(repo.diffs) != 1 || repo.diffs[0] != 1 || len(repo.inserted) != 1 || repo.inserted[0].Price != 98.75 {
		t.Fatalf("unexpected add operations: %+v", repo)
	}
	repo.updated, repo.diffs = nil, nil
	remove := []model.BondOperation{{Name: "A", Code: "A", GroupName: "G", LotAmount: 1, LotSize: 10, Price: 101.25}, {Name: "B", Code: "B", GroupName: "G", LotAmount: 3, LotSize: 1, Price: 99.5}}
	if err := a.RemoveBonds(context.Background(), 1, remove); err != nil {
		t.Fatal(err)
	}
	if len(repo.updated) != 1 || repo.updated[0].Price != 101.25 || repo.updated[0].LotAmount != 2 || len(repo.diffs) != 1 || repo.diffs[0] != -1 || len(repo.removed) != 1 || repo.removed[0].Price != 99.5 || repo.removed[0].LotAmount != 2 {
		t.Fatalf("unexpected remove operations: %+v", repo)
	}
}

type bondTestMoex struct {
	MoexApi
	info []model.MoexBondInfoResult
	find []model.MoexBondFindResult
}

func (m bondTestMoex) GetBondsInfo(context.Context, []string) []model.MoexBondInfoResult {
	return m.info
}
func (m bondTestMoex) FindBonds(context.Context, []string) []model.MoexBondFindResult { return m.find }

func TestBondValidation(t *testing.T) {
	for _, remove := range []bool{false, true} {
		for _, tc := range []struct {
			bonds []model.BondOperation
			want  error
		}{
			{[]model.BondOperation{{Name: " ", Code: "A", LotAmount: 1}}, servererrors.ErrInvalidBondName},
			{[]model.BondOperation{{Name: "A", Code: "A", LotAmount: 0}}, servererrors.ErrWrongBondAmount},
			{[]model.BondOperation{{Name: "A", Code: "A", LotAmount: -1}}, servererrors.ErrWrongBondAmount},
			{[]model.BondOperation{{Name: "A", Code: "A", LotAmount: 1}}, servererrors.ErrInvalidBondPrice},
			{[]model.BondOperation{{Name: "A", Code: "A", LotAmount: 1, Price: -1}}, servererrors.ErrInvalidBondPrice},
			{[]model.BondOperation{{Name: "A", Code: "A", GroupName: "G", LotAmount: 1, Price: 1}, {Name: "A", Code: "A", GroupName: "G", LotAmount: 1, Price: 1}}, servererrors.ErrDuplicateBond},
			{[]model.BondOperation{{Name: "A", Code: "A", GroupName: "other", LotAmount: 1, Price: 1}}, servererrors.ErrWrongBondGroup},
		} {
			repo := &bondTestRepo{bonds: []model.Bond{{Code: "A", GroupName: "G", LotAmount: 2}}}
			a := New(repo, nil, nil)
			var err error
			if remove {
				err = a.RemoveBonds(context.Background(), 1, tc.bonds)
			} else {
				err = a.AddBonds(context.Background(), 1, tc.bonds)
			}
			if !errors.Is(err, tc.want) || repo.writes != 0 {
				t.Fatalf("remove=%v: error %v, writes %d, want %v and no writes", remove, err, repo.writes, tc.want)
			}
		}
	}
}

func TestBondMoexEnrichmentAndFailures(t *testing.T) {
	repo := &bondTestRepo{bonds: []model.Bond{{Code: "A", Name: "stored", GroupName: "G", LotAmount: 3}}}
	market := model.Bond{Code: "A", GroupName: "wrong", LotAmount: 99, CleanPrice: 42, LotSize: 2}
	moex := bondTestMoex{info: []model.MoexBondInfoResult{{Bond: market}}, find: []model.MoexBondFindResult{{SearchResults: []model.MoexBondSearchResult{{Bond: market}}}}}
	a := New(repo, nil, moex)
	bonds, err := a.GetBonds(context.Background(), 1)
	if err != nil || len(bonds) != 1 || bonds[0].GroupName != "G" || bonds[0].LotAmount != 3 || bonds[0].CleanPrice != 42 {
		t.Fatalf("GetBonds: %v, %v", bonds, err)
	}
	found, err := a.FindBonds(context.Background(), 1, []string{"A"})
	if err != nil || len(found) != 1 || len(found[0].SearchResults) != 1 || found[0].SearchResults[0].GroupName != "G" || found[0].SearchResults[0].LotAmount != 3 {
		t.Fatalf("FindBonds: %v, %v", found, err)
	}
	moex.info[0].Err = errors.New("failure")
	moex.find[0].SearchResults[0].Err = errors.New("failure")
	a = New(repo, nil, moex)
	if _, err = a.GetBonds(context.Background(), 1); !errors.Is(err, servererrors.ErrMoexUnhandled) {
		t.Fatalf("GetBonds error: %v", err)
	}
	if _, err = a.FindBonds(context.Background(), 1, []string{"A"}); !errors.Is(err, servererrors.ErrMoexUnhandled) {
		t.Fatalf("FindBonds error: %v", err)
	}
}

func TestBondWeights(t *testing.T) {
	repo := &bondTestRepo{}
	a := New(repo, nil, nil)
	for _, groups := range [][]model.BondGroup{{{Weight: -1}}, {{Weight: 101}}, {{Weight: 60}, {Weight: 41}}} {
		if err := a.UpdateBondWeights(context.Background(), 1, groups); !errors.Is(err, servererrors.ErrInvalidWeights) {
			t.Fatalf("error: %v", err)
		}
	}
	if err := a.UpdateBondWeights(context.Background(), 1, []model.BondGroup{{Weight: 40}}); err != nil || repo.writes != 1 {
		t.Fatalf("error: %v, writes: %d", err, repo.writes)
	}
}
