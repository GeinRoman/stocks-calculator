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
	bonds  []model.Bond
	writes int
}

func (r *bondTestRepo) GetBonds(context.Context, int) ([]model.Bond, error) {
	return append([]model.Bond(nil), r.bonds...), nil
}
func (r *bondTestRepo) InsertBonds(context.Context, int, []model.Bond) error { r.writes++; return nil }
func (r *bondTestRepo) UpdateBondsAmount(context.Context, int, []model.Bond, []int) error {
	r.writes++
	return nil
}
func (r *bondTestRepo) RemoveBonds(context.Context, int, []model.Bond) error { r.writes++; return nil }
func (r *bondTestRepo) UpdateBondWeights(context.Context, int, []model.BondGroup) error {
	r.writes++
	return nil
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
			bonds []model.Bond
			want  error
		}{
			{[]model.Bond{{Code: "A", LotAmount: 0}}, servererrors.ErrWrongBondAmount},
			{[]model.Bond{{Code: "A", LotAmount: -1}}, servererrors.ErrWrongBondAmount},
			{[]model.Bond{{Code: "A", GroupName: "G", LotAmount: 1}, {Code: "A", GroupName: "G", LotAmount: 1}}, servererrors.ErrDuplicateBond},
			{[]model.Bond{{Code: "A", GroupName: "other", LotAmount: 1}}, servererrors.ErrWrongBondGroup},
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
