package app

import (
	"context"
	"errors"
	"math"
	"testing"

	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/servererrors"
)

type contractTestRepo struct {
	Repository
	calls int
}

func (r *contractTestRepo) AddDeposit(context.Context, int, model.Deposit) error {
	r.calls++
	return nil
}
func (r *contractTestRepo) UpdateDeposit(context.Context, int, model.DepositUpdate) error {
	r.calls++
	return nil
}
func (r *contractTestRepo) RemoveDeposit(context.Context, int, int64) error { r.calls++; return nil }
func (r *contractTestRepo) CloseDeposit(context.Context, int, model.DepositClose) error {
	r.calls++
	return nil
}
func (r *contractTestRepo) AddInsurance(context.Context, int, model.Insurance) error {
	r.calls++
	return nil
}
func (r *contractTestRepo) UpdateInsurance(context.Context, int, model.InsuranceUpdate) error {
	r.calls++
	return nil
}
func (r *contractTestRepo) RemoveInsurance(context.Context, int, int64) error { r.calls++; return nil }
func (r *contractTestRepo) CloseInsurance(context.Context, int, model.InsuranceClose) error {
	r.calls++
	return nil
}

func TestContractValidation(t *testing.T) {
	ctx := context.Background()
	r := &contractTestRepo{}
	s := &app{repo: r}
	d := model.Deposit{Name: "deposit", BankName: "bank", AmountRUB: 100, EndDate: "2030-02-28"}
	i := model.Insurance{Name: "insurance", CompanyName: "company", AmountRUB: 100, EndDate: "2030-02-28"}
	if !errors.Is(s.AddDeposit(ctx, 1, model.Deposit{}), servererrors.ErrInvalidContract) || !errors.Is(s.AddInsurance(ctx, 1, model.Insurance{}), servererrors.ErrInvalidContract) {
		t.Fatal("invalid input must return ErrInvalidContract")
	}
	for _, amount := range []float64{0, -1, math.NaN(), math.Inf(1), math.Inf(-1)} {
		badD, badI := d, i
		badD.AmountRUB, badI.AmountRUB = amount, amount
		if s.AddDeposit(ctx, 1, badD) == nil || s.AddInsurance(ctx, 1, badI) == nil {
			t.Errorf("accepted amount %v", amount)
		}
		if s.CloseDeposit(ctx, 1, model.DepositClose{ID: 1, ClosingAmountRUB: amount}) == nil || s.CloseInsurance(ctx, 1, model.InsuranceClose{ID: 1, ClosingAmountRUB: amount}) == nil {
			t.Errorf("accepted closing amount %v", amount)
		}
	}
	for _, end := range []string{"", "2030-02-29", "2030-2-01", "0000-01-01", "2030-01-01T00:00:00Z"} {
		badD, badI := d, i
		badD.EndDate, badI.EndDate = end, end
		if s.AddDeposit(ctx, 1, badD) == nil || s.AddInsurance(ctx, 1, badI) == nil {
			t.Errorf("accepted date %q", end)
		}
	}
	for _, rate := range []float64{-1, math.NaN(), math.Inf(1), math.Inf(-1)} {
		if s.UpdateDeposit(ctx, 1, model.DepositUpdate{ID: 1, AnnualInterest: &rate}) == nil || s.UpdateInsurance(ctx, 1, model.InsuranceUpdate{ID: 1, AnnualInterest: &rate}) == nil {
			t.Errorf("accepted rate %v", rate)
		}
	}
	if s.AddDeposit(ctx, 1, model.Deposit{Name: " ", BankName: "bank", AmountRUB: 1, EndDate: d.EndDate}) == nil || s.AddInsurance(ctx, 1, model.Insurance{Name: "x", CompanyName: " ", AmountRUB: 1, EndDate: i.EndDate}) == nil {
		t.Error("accepted blank name or provider")
	}
	if s.RemoveDeposit(ctx, 1, 0) == nil || s.RemoveInsurance(ctx, 1, -1) == nil || s.CloseDeposit(ctx, 1, model.DepositClose{ClosingAmountRUB: 1}) == nil || s.CloseInsurance(ctx, 1, model.InsuranceClose{ClosingAmountRUB: 1}) == nil {
		t.Error("accepted invalid id")
	}
	if r.calls != 0 {
		t.Fatalf("invalid inputs reached repository %d times", r.calls)
	}
	end := "2028-02-29"
	if s.AddDeposit(ctx, 1, d) != nil || s.AddInsurance(ctx, 1, i) != nil || s.UpdateDeposit(ctx, 1, model.DepositUpdate{ID: 1, EndDate: &end}) != nil || s.UpdateInsurance(ctx, 1, model.InsuranceUpdate{ID: 1, EndDate: &i.EndDate}) != nil || s.CloseDeposit(ctx, 1, model.DepositClose{ID: 1, ClosingAmountRUB: 1}) != nil || s.CloseInsurance(ctx, 1, model.InsuranceClose{ID: 1, ClosingAmountRUB: 1}) != nil {
		t.Error("rejected valid input")
	}
	if r.calls != 6 {
		t.Errorf("expected 6 repository calls, got %d", r.calls)
	}
}

func ptr[T any](v T) *T { return &v }

func TestContractPartialUpdates(t *testing.T) {
	ctx := context.Background()
	r := &contractTestRepo{}
	s := &app{repo: r}
	for _, tc := range []struct {
		name  string
		d     model.DepositUpdate
		i     model.InsuranceUpdate
		valid bool
	}{
		{"name", model.DepositUpdate{ID: 1, Name: ptr("new")}, model.InsuranceUpdate{ID: 1, Name: ptr("new")}, true},
		{"provider", model.DepositUpdate{ID: 1, BankName: ptr("bank")}, model.InsuranceUpdate{ID: 1, CompanyName: ptr("company")}, true},
		{"zero rate", model.DepositUpdate{ID: 1, AnnualInterest: ptr(0.0)}, model.InsuranceUpdate{ID: 1, AnnualInterest: ptr(0.0)}, true},
		{"date", model.DepositUpdate{ID: 1, EndDate: ptr("2028-02-29")}, model.InsuranceUpdate{ID: 1, EndDate: ptr("2028-02-29")}, true},
		{"empty", model.DepositUpdate{ID: 1}, model.InsuranceUpdate{ID: 1}, false},
		{"missing id", model.DepositUpdate{Name: ptr("new")}, model.InsuranceUpdate{Name: ptr("new")}, false},
		{"blank name", model.DepositUpdate{ID: 1, Name: ptr(" ")}, model.InsuranceUpdate{ID: 1, Name: ptr(" ")}, false},
		{"blank provider", model.DepositUpdate{ID: 1, BankName: ptr("")}, model.InsuranceUpdate{ID: 1, CompanyName: ptr("")}, false},
		{"invalid date", model.DepositUpdate{ID: 1, EndDate: ptr("2023-02-29")}, model.InsuranceUpdate{ID: 1, EndDate: ptr("2023-02-29")}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			before := r.calls
			ed, ei := s.UpdateDeposit(ctx, 7, tc.d), s.UpdateInsurance(ctx, 7, tc.i)
			if tc.valid && (ed != nil || ei != nil || r.calls != before+2) || !tc.valid && (!errors.Is(ed, servererrors.ErrInvalidContract) || !errors.Is(ei, servererrors.ErrInvalidContract) || r.calls != before) {
				t.Fatalf("deposit: %v, insurance: %v, calls: %d", ed, ei, r.calls-before)
			}
		})
	}
}
