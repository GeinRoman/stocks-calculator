package repo

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/servererrors"
)

func TestCreditCardRepositoryIntegration(t *testing.T) {
	dsn := os.Getenv("CREDIT_CARD_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set CREDIT_CARD_TEST_DATABASE_URL to a disposable PostgreSQL database")
	}
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		var err error
		dsn, err = pq.ParseURL(dsn)
		require.NoError(t, err)
	}
	admin, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = admin.Close() })
	var random [16]byte
	_, err = rand.Read(random[:])
	require.NoError(t, err)
	schema := fmt.Sprintf("credit_card_test_%x", random)
	_, err = admin.Exec("CREATE SCHEMA " + pq.QuoteIdentifier(schema))
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := admin.Exec("DROP SCHEMA " + pq.QuoteIdentifier(schema) + " CASCADE")
		require.NoError(t, err)
	})
	db, err := sql.Open("postgres", dsn+" search_path="+schema)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	_, err = db.Exec(`CREATE TABLE profiles (id SERIAL PRIMARY KEY, user_id INTEGER NOT NULL, is_default BOOLEAN NOT NULL);
		CREATE TABLE credit_cards (id BIGSERIAL PRIMARY KEY, profile_id INTEGER NOT NULL REFERENCES profiles(id) ON DELETE CASCADE, name TEXT NOT NULL, bank_name TEXT NOT NULL, mandatory_payment_date DATE NOT NULL, mandatory_payment_amount_rub DOUBLE PRECISION NOT NULL, overall_limit_rub DOUBLE PRECISION NOT NULL);
		INSERT INTO profiles (user_id,is_default) VALUES (1,true),(2,true),(1,false);`)
	require.NoError(t, err)
	ctx := context.Background()
	r := New(db)
	card := model.CreditCard{Name: "card", BankName: "bank", MandatoryPaymentDate: "2024-02-29", MandatoryPaymentAmountRUB: 1.5, OverallLimitRUB: 100}
	require.NoError(t, r.AddCreditCard(ctx, 1, card))
	cards, err := r.GetCreditCards(ctx, 1)
	require.NoError(t, err)
	require.Len(t, cards, 1)
	card.ID = cards[0].ID
	require.Equal(t, card, cards[0])
	other, err := r.GetCreditCards(ctx, 2)
	require.NoError(t, err)
	require.Empty(t, other)
	require.ErrorIs(t, r.UpdateCreditCard(ctx, 2, model.CreditCardUpdate{ID: card.ID, Name: personalAssetPtr("updated")}), model.ErrCreditCardNotFound)
	require.ErrorIs(t, r.RemoveCreditCard(ctx, 2, card.ID), model.ErrCreditCardNotFound)
	card.Name = "updated"
	require.NoError(t, r.UpdateCreditCard(ctx, 1, model.CreditCardUpdate{ID: card.ID, Name: &card.Name}))
	cards, err = r.GetCreditCards(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, card, cards[0])
	require.NoError(t, r.UpdateCreditCard(ctx, 1, model.CreditCardUpdate{ID: card.ID, MandatoryPaymentAmountRUB: personalAssetPtr(0.0), OverallLimitRUB: personalAssetPtr(0.0)}))
	card.MandatoryPaymentAmountRUB, card.OverallLimitRUB = 0, 0
	cards, err = r.GetCreditCards(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, card, cards[0])
	newDate := "2025-03-01"
	require.NoError(t, r.UpdateCreditCard(ctx, 1, model.CreditCardUpdate{ID: card.ID, MandatoryPaymentDate: &newDate}))
	card.MandatoryPaymentDate = newDate
	cards, err = r.GetCreditCards(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, card, cards[0])
	require.NoError(t, r.RemoveCreditCard(ctx, 1, card.ID))
	require.ErrorIs(t, r.RemoveCreditCard(ctx, 1, card.ID), model.ErrCreditCardNotFound)
	_, err = r.GetCreditCards(ctx, 3)
	require.ErrorIs(t, err, servererrors.ErrNoDefaultProfile)
}
