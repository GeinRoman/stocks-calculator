package app

import (
	"context"
	"errors"
	"math"
	"testing"

	"stocks_calculator/internal/model"
)

type creditCardTestRepo struct {
	Repository
	called string
	user   int
	card   model.CreditCard
	update model.CreditCardUpdate
	id     int64
}

func cardUpdate(c model.CreditCard) model.CreditCardUpdate {
	return model.CreditCardUpdate{ID: c.ID, Name: &c.Name, BankName: &c.BankName, MandatoryPaymentDate: &c.MandatoryPaymentDate, MandatoryPaymentAmountRUB: &c.MandatoryPaymentAmountRUB, OverallLimitRUB: &c.OverallLimitRUB}
}

func (r *creditCardTestRepo) GetCreditCards(_ context.Context, user int) ([]model.CreditCard, error) {
	r.called, r.user = "get", user
	return []model.CreditCard{{ID: 1}}, nil
}
func (r *creditCardTestRepo) AddCreditCard(_ context.Context, user int, card model.CreditCard) error {
	r.called, r.user, r.card = "add", user, card
	return nil
}
func (r *creditCardTestRepo) UpdateCreditCard(_ context.Context, user int, card model.CreditCardUpdate) error {
	r.called, r.user, r.update = "update", user, card
	return nil
}
func (r *creditCardTestRepo) RemoveCreditCard(_ context.Context, user int, id int64) error {
	r.called, r.user, r.id = "remove", user, id
	return nil
}

func TestCreditCardValidation(t *testing.T) {
	valid := model.CreditCard{Name: "card", BankName: "bank", MandatoryPaymentDate: "2024-02-29"}
	tests := []struct {
		name   string
		change func(*model.CreditCard)
	}{
		{"name", func(c *model.CreditCard) { c.Name = " " }},
		{"bank", func(c *model.CreditCard) { c.BankName = "" }},
		{"date", func(c *model.CreditCard) { c.MandatoryPaymentDate = "2023-02-29" }},
		{"date format", func(c *model.CreditCard) { c.MandatoryPaymentDate = "2024-2-29" }},
		{"zero year", func(c *model.CreditCard) { c.MandatoryPaymentDate = "0000-01-01" }},
		{"negative amount", func(c *model.CreditCard) { c.MandatoryPaymentAmountRUB = -1 }},
		{"nan amount", func(c *model.CreditCard) { c.MandatoryPaymentAmountRUB = math.NaN() }},
		{"infinite amount", func(c *model.CreditCard) { c.MandatoryPaymentAmountRUB = math.Inf(1) }},
		{"negative limit", func(c *model.CreditCard) { c.OverallLimitRUB = -1 }},
		{"nan limit", func(c *model.CreditCard) { c.OverallLimitRUB = math.NaN() }},
		{"infinite limit", func(c *model.CreditCard) { c.OverallLimitRUB = math.Inf(1) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &creditCardTestRepo{}
			card := valid
			tt.change(&card)
			if err := New(r, nil, nil).AddCreditCard(context.Background(), 1, card); !errors.Is(err, model.ErrInvalidCreditCard) || r.called != "" {
				t.Fatalf("invalid card: %v, called %q", err, r.called)
			}
			card.ID = 1
			if err := New(r, nil, nil).UpdateCreditCard(context.Background(), 1, cardUpdate(card)); !errors.Is(err, model.ErrInvalidCreditCard) || r.called != "" {
				t.Fatalf("invalid update: %v, called %q", err, r.called)
			}
		})
	}
	r := &creditCardTestRepo{}
	s := New(r, nil, nil)
	ctx := context.Background()
	if err := s.AddCreditCard(ctx, 7, valid); err != nil || r.called != "add" || r.user != 7 || r.card != valid {
		t.Fatalf("add: %v", err)
	}
	valid.ID = 3
	if err := s.AddCreditCard(ctx, 7, valid); !errors.Is(err, model.ErrInvalidCreditCard) {
		t.Fatalf("add id: %v", err)
	}
	update := cardUpdate(valid)
	if err := s.UpdateCreditCard(ctx, 7, update); err != nil || r.called != "update" || r.update != update {
		t.Fatalf("update: %v", err)
	}
	valid.ID = 0
	if err := s.UpdateCreditCard(ctx, 7, cardUpdate(valid)); !errors.Is(err, model.ErrInvalidCreditCard) {
		t.Fatalf("update id: %v", err)
	}
	if err := s.RemoveCreditCard(ctx, 7, 0); !errors.Is(err, model.ErrInvalidCreditCard) {
		t.Fatalf("remove id: %v", err)
	}
	if err := s.RemoveCreditCard(ctx, 7, 3); err != nil || r.called != "remove" || r.id != 3 {
		t.Fatalf("remove: %v", err)
	}
	cards, err := s.GetCreditCards(ctx, 7)
	if err != nil || r.called != "get" || len(cards) != 1 {
		t.Fatalf("get: %v", err)
	}
}

func TestCreditCardPartialUpdates(t *testing.T) {
	for _, tc := range []struct {
		name   string
		update model.CreditCardUpdate
		valid  bool
	}{
		{"name", model.CreditCardUpdate{ID: 1, Name: ptr("new")}, true},
		{"bank", model.CreditCardUpdate{ID: 1, BankName: ptr("bank")}, true},
		{"date", model.CreditCardUpdate{ID: 1, MandatoryPaymentDate: ptr("2024-02-29")}, true},
		{"zero payment", model.CreditCardUpdate{ID: 1, MandatoryPaymentAmountRUB: ptr(0.0)}, true},
		{"zero limit", model.CreditCardUpdate{ID: 1, OverallLimitRUB: ptr(0.0)}, true},
		{"empty", model.CreditCardUpdate{ID: 1}, false},
		{"missing id", model.CreditCardUpdate{Name: ptr("new")}, false},
		{"blank name", model.CreditCardUpdate{ID: 1, Name: ptr(" ")}, false},
		{"blank bank", model.CreditCardUpdate{ID: 1, BankName: ptr("")}, false},
		{"bad date", model.CreditCardUpdate{ID: 1, MandatoryPaymentDate: ptr("2023-02-29")}, false},
		{"negative payment", model.CreditCardUpdate{ID: 1, MandatoryPaymentAmountRUB: ptr(-1.0)}, false},
		{"nonfinite limit", model.CreditCardUpdate{ID: 1, OverallLimitRUB: ptr(math.Inf(1))}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := &creditCardTestRepo{}
			err := New(r, nil, nil).UpdateCreditCard(context.Background(), 7, tc.update)
			if tc.valid && (err != nil || r.called != "update" || r.user != 7 || r.update != tc.update) || !tc.valid && (!errors.Is(err, model.ErrInvalidCreditCard) || r.called != "") {
				t.Fatalf("error: %v, repository call: %q", err, r.called)
			}
		})
	}
}
