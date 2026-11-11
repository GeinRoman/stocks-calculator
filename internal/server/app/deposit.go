package app

import (
	"context"
	"math"
	"strings"
	"time"

	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/servererrors"
)

func validDepositMetadata(name, provider string, rate float64, end string) bool {
	date, err := time.Parse("2006-01-02", end)
	return strings.TrimSpace(name) != "" && strings.TrimSpace(provider) != "" && rate >= 0 && !math.IsNaN(rate) && !math.IsInf(rate, 0) && err == nil && date.Year() >= 1 && date.Format("2006-01-02") == end
}

func validDepositAmount(amount float64) bool {
	return amount > 0 && !math.IsNaN(amount) && !math.IsInf(amount, 0)
}

func validOptionalName(s *string) bool {
	return s == nil || strings.TrimSpace(*s) != ""
}

func validOptionalNumber(n *float64) bool {
	return n == nil || (*n >= 0 && !math.IsNaN(*n) && !math.IsInf(*n, 0))
}

func validOptionalDate(s *string) bool {
	if s == nil {
		return true
	}
	date, err := time.Parse("2006-01-02", *s)
	return err == nil && date.Year() >= 1 && date.Format("2006-01-02") == *s
}

func (s *app) GetDeposits(ctx context.Context, user int) ([]model.Deposit, error) {
	return s.repo.GetDeposits(ctx, user)
}

func (s *app) AddDeposit(ctx context.Context, user int, d model.Deposit) error {
	if !validDepositMetadata(d.Name, d.BankName, d.AnnualInterest, d.EndDate) || !validDepositAmount(d.AmountRUB) {
		return servererrors.ErrInvalidContract
	}
	return s.repo.AddDeposit(ctx, user, d)
}

func (s *app) UpdateDeposit(ctx context.Context, user int, d model.DepositUpdate) error {
	if d.ID <= 0 || (d.Name == nil && d.BankName == nil && d.AnnualInterest == nil && d.EndDate == nil) ||
		!validOptionalName(d.Name) || !validOptionalName(d.BankName) || !validOptionalNumber(d.AnnualInterest) || !validOptionalDate(d.EndDate) {
		return servererrors.ErrInvalidContract
	}
	return s.repo.UpdateDeposit(ctx, user, d)
}

func (s *app) RemoveDeposit(ctx context.Context, user int, id int64) error {
	if id <= 0 {
		return servererrors.ErrInvalidContract
	}
	return s.repo.RemoveDeposit(ctx, user, id)
}

func (s *app) CloseDeposit(ctx context.Context, user int, d model.DepositClose) error {
	if d.ID <= 0 || !validDepositAmount(d.ClosingAmountRUB) {
		return servererrors.ErrInvalidContract
	}
	return s.repo.CloseDeposit(ctx, user, d)
}
