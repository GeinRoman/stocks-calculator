package app

import (
	"context"
	"math"
	"strings"
	"time"

	"stocks_calculator/internal/model"
)

func (s *app) GetCreditCards(ctx context.Context, user int) ([]model.CreditCard, error) {
	return s.repo.GetCreditCards(ctx, user)
}

func validCreditCard(c model.CreditCard) bool {
	date, err := time.Parse("2006-01-02", c.MandatoryPaymentDate)
	return strings.TrimSpace(c.Name) != "" && strings.TrimSpace(c.BankName) != "" &&
		err == nil && date.Year() >= 1 && date.Format("2006-01-02") == c.MandatoryPaymentDate &&
		c.MandatoryPaymentAmountRUB >= 0 && !math.IsNaN(c.MandatoryPaymentAmountRUB) && !math.IsInf(c.MandatoryPaymentAmountRUB, 0) &&
		c.OverallLimitRUB >= 0 && !math.IsNaN(c.OverallLimitRUB) && !math.IsInf(c.OverallLimitRUB, 0)
}

func (s *app) AddCreditCard(ctx context.Context, user int, card model.CreditCard) error {
	if card.ID != 0 || !validCreditCard(card) {
		return model.ErrInvalidCreditCard
	}
	return s.repo.AddCreditCard(ctx, user, card)
}

func (s *app) UpdateCreditCard(ctx context.Context, user int, card model.CreditCardUpdate) error {
	if card.ID <= 0 || (card.Name == nil && card.BankName == nil && card.MandatoryPaymentDate == nil && card.MandatoryPaymentAmountRUB == nil && card.OverallLimitRUB == nil) ||
		!validOptionalName(card.Name) || !validOptionalName(card.BankName) || !validOptionalDate(card.MandatoryPaymentDate) ||
		!validOptionalNumber(card.MandatoryPaymentAmountRUB) || !validOptionalNumber(card.OverallLimitRUB) {
		return model.ErrInvalidCreditCard
	}
	return s.repo.UpdateCreditCard(ctx, user, card)
}

func (s *app) RemoveCreditCard(ctx context.Context, user int, id int64) error {
	if id <= 0 {
		return model.ErrInvalidCreditCard
	}
	return s.repo.RemoveCreditCard(ctx, user, id)
}
