package app

import (
	"context"
	"errors"
	"math"
	"strings"
	"testing"

	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/servererrors"
)

type cryptoTestRepo struct {
	Repository
	called  bool
	ops     []model.CryptoOperation
	updated model.CryptoAssetUpdate
}

func (r *cryptoTestRepo) TransactCrypto(_ context.Context, _ int, ops []model.CryptoOperation, _ bool) error {
	r.called = true
	r.ops = ops
	return nil
}
func (r *cryptoTestRepo) UpdateCryptoAsset(_ context.Context, _ int, asset model.CryptoAssetUpdate) error {
	r.called = true
	r.updated = asset
	return nil
}

func TestCryptoTransactionValidation(t *testing.T) {
	valid := model.CryptoOperation{CryptoAsset: model.CryptoAsset{Name: "BTC", GroupName: "coins", Platform: "exchange", Amount: 0.25}, Price: model.Money{Price: 1, Currency: "USDT"}}
	tests := []struct {
		name   string
		mutate func(*model.CryptoOperation)
		valid  bool
	}{
		{"fractional", func(*model.CryptoOperation) {}, true},
		{"negative interest", func(o *model.CryptoOperation) { o.AnnualInterest = -1 }, false},
		{"nan interest", func(o *model.CryptoOperation) { o.AnnualInterest = math.NaN() }, false},
		{"zero amount", func(o *model.CryptoOperation) { o.Amount = 0 }, false},
		{"infinite amount", func(o *model.CryptoOperation) { o.Amount = math.Inf(1) }, false},
		{"zero price", func(o *model.CryptoOperation) { o.Price.Price = 0 }, false},
		{"long currency", func(o *model.CryptoOperation) { o.Price.Currency = strings.Repeat("x", 17) }, false},
		{"max currency", func(o *model.CryptoOperation) { o.Price.Currency = strings.Repeat("x", 16) }, true},
		{"empty platform", func(o *model.CryptoOperation) { o.Platform = " " }, false},
		{"empty group", func(o *model.CryptoOperation) { o.GroupName = " " }, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &cryptoTestRepo{}
			s := New(r, nil, nil)
			op := valid
			tt.mutate(&op)
			err := s.BuyCrypto(context.Background(), 1, []model.CryptoOperation{op})
			if tt.valid && (err != nil || !r.called) {
				t.Fatalf("expected success, got %v", err)
			}
			if !tt.valid && (!errors.Is(err, servererrors.ErrInvalidCrypto) || r.called) {
				t.Fatalf("expected validation failure, got %v", err)
			}
		})
	}
	r := &cryptoTestRepo{}
	valid.ID = 7
	if err := New(r, nil, nil).SellCrypto(context.Background(), 1, []model.CryptoOperation{valid, valid}); !errors.Is(err, servererrors.ErrInvalidCrypto) || r.called {
		t.Fatalf("duplicate: %v", err)
	}
}

func TestCryptoIDOperationsAndUpdate(t *testing.T) {
	ctx := context.Background()
	r := &cryptoTestRepo{}
	s := New(r, nil, nil)
	op := model.CryptoOperation{CryptoAsset: model.CryptoAsset{ID: 42, Amount: 0.5}, Price: model.Money{Price: 2, Currency: "USD"}}
	if err := s.BuyCrypto(ctx, 1, []model.CryptoOperation{op}); err != nil || !r.called {
		t.Fatalf("buy by ID: %v", err)
	}
	r.called = false
	if err := s.SellCrypto(ctx, 1, []model.CryptoOperation{op}); err != nil || !r.called {
		t.Fatalf("sell by ID: %v", err)
	}
	update := model.CryptoAssetUpdate{ID: 42, Name: "renamed", Platform: "wallet", URL: "https://example.com", AnnualInterest: 2}
	r.called = false
	if err := s.UpdateCryptoAsset(ctx, 1, update); err != nil || !r.called || r.updated != update {
		t.Fatalf("update: %v", err)
	}
	for _, invalid := range []model.CryptoAssetUpdate{{}, {ID: 1, Name: "x", Platform: "p", AnnualInterest: math.Inf(1)}, {ID: 1, Name: "x", Platform: "p", AnnualInterest: -1}} {
		r.called = false
		if err := s.UpdateCryptoAsset(ctx, 1, invalid); !errors.Is(err, servererrors.ErrInvalidCrypto) || r.called {
			t.Fatalf("invalid update: %v", err)
		}
	}
	op.ID = 0
	r.called = false
	if err := s.SellCrypto(ctx, 1, []model.CryptoOperation{op}); !errors.Is(err, servererrors.ErrInvalidCrypto) || r.called {
		t.Fatalf("sell without ID: %v", err)
	}
}
