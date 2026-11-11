package app

import (
	"context"
	"math"
	"strings"
	"time"

	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/servererrors"
)

func validInsuranceMetadata(name, company string, rate float64, end string) bool {
	date, err := time.Parse("2006-01-02", end)
	return strings.TrimSpace(name) != "" && strings.TrimSpace(company) != "" && rate >= 0 && !math.IsNaN(rate) && !math.IsInf(rate, 0) && err == nil && date.Year() >= 1 && date.Format("2006-01-02") == end
}

func validInsuranceAmount(amount float64) bool {
	return amount > 0 && !math.IsNaN(amount) && !math.IsInf(amount, 0)
}

func (s *app) GetInsurances(ctx context.Context, user int) ([]model.Insurance, error) {
	return s.repo.GetInsurances(ctx, user)
}

func (s *app) AddInsurance(ctx context.Context, user int, d model.Insurance) error {
	if !validInsuranceMetadata(d.Name, d.CompanyName, d.AnnualInterest, d.EndDate) || !validInsuranceAmount(d.AmountRUB) {
		return servererrors.ErrInvalidContract
	}
	return s.repo.AddInsurance(ctx, user, d)
}

func (s *app) UpdateInsurance(ctx context.Context, user int, d model.InsuranceUpdate) error {
	if d.ID <= 0 || (d.Name == nil && d.CompanyName == nil && d.AnnualInterest == nil && d.EndDate == nil) ||
		!validOptionalName(d.Name) || !validOptionalName(d.CompanyName) || !validOptionalNumber(d.AnnualInterest) || !validOptionalDate(d.EndDate) {
		return servererrors.ErrInvalidContract
	}
	return s.repo.UpdateInsurance(ctx, user, d)
}

func (s *app) RemoveInsurance(ctx context.Context, user int, id int64) error {
	if id <= 0 {
		return servererrors.ErrInvalidContract
	}
	return s.repo.RemoveInsurance(ctx, user, id)
}

func (s *app) CloseInsurance(ctx context.Context, user int, d model.InsuranceClose) error {
	if d.ID <= 0 || !validInsuranceAmount(d.ClosingAmountRUB) {
		return servererrors.ErrInvalidContract
	}
	return s.repo.CloseInsurance(ctx, user, d)
}
