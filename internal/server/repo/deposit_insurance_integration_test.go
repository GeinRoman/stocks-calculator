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

func personalAssetPtr[T any](v T) *T { return &v }

// PERSONAL_ASSETS_TEST_DATABASE_URL must point to a disposable PostgreSQL database.
func TestDepositInsuranceIntegration(t *testing.T) {
	dsn := os.Getenv("PERSONAL_ASSETS_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set PERSONAL_ASSETS_TEST_DATABASE_URL to run PostgreSQL integration tests")
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
	schema := fmt.Sprintf("personal_assets_test_%x", random)
	_, err = admin.Exec("CREATE SCHEMA " + pq.QuoteIdentifier(schema))
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := admin.Exec("DROP SCHEMA " + pq.QuoteIdentifier(schema) + " CASCADE")
		require.NoError(t, err)
	})
	db, err := sql.Open("postgres", dsn+" search_path="+schema)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	_, err = db.Exec(`CREATE TABLE profiles (id SERIAL PRIMARY KEY, user_id INTEGER NOT NULL, is_default BOOLEAN NOT NULL); INSERT INTO profiles (user_id,is_default) VALUES (1,true),(2,true),(1,false);`)
	require.NoError(t, err)
	migration, err := os.ReadFile("../../../migrations/000011_create_deposits_insurances_credit_cards.up.sql")
	require.NoError(t, err)
	_, err = db.Exec(string(migration))
	require.NoError(t, err)
	_, err = db.Exec(`
		INSERT INTO deposit_transactions (profile_id,name,bank_name,amount_rub,buying) VALUES (1,'legacy','bank',1,true);
		INSERT INTO deposits (profile_id,acquisition_transaction_id,name,bank_name,amount_rub,annual_interest,end_date)
		SELECT 1,id,'legacy','bank',1,0,'2030-01-01' FROM deposit_transactions WHERE name='legacy';
		INSERT INTO insurance_transactions (profile_id,name,company_name,amount_rub,buying) VALUES (1,'legacy','company',1,true);
		INSERT INTO insurances (profile_id,acquisition_transaction_id,name,company_name,amount_rub,annual_interest,end_date)
		SELECT 1,id,'legacy','company',1,0,'2030-01-01' FROM insurance_transactions WHERE name='legacy';`)
	require.NoError(t, err)
	migration, err = os.ReadFile("../../../migrations/000012_add_contract_start_dates.up.sql")
	require.NoError(t, err)
	_, err = db.Exec(string(migration))
	require.NoError(t, err)
	ctx := context.Background()
	r := New(db)
	legacyDeposits, err := r.GetDeposits(ctx, 1)
	require.NoError(t, err)
	require.Len(t, legacyDeposits, 1)
	require.Nil(t, legacyDeposits[0].StartDate)
	legacyInsurances, err := r.GetInsurances(ctx, 1)
	require.NoError(t, err)
	require.Len(t, legacyInsurances, 1)
	require.Nil(t, legacyInsurances[0].StartDate)
	require.NoError(t, r.RemoveDeposit(ctx, 1, legacyDeposits[0].ID))
	require.NoError(t, r.RemoveInsurance(ctx, 1, legacyInsurances[0].ID))

	t.Run("start date round trip and ordering", func(t *testing.T) {
		start := "2020-01-01"
		require.NoError(t, r.AddDeposit(ctx, 1, model.Deposit{Name: "date", BankName: "bank", AmountRUB: 1, StartDate: &start, EndDate: "2030-01-01"}))
		require.NoError(t, r.AddInsurance(ctx, 1, model.Insurance{Name: "date", CompanyName: "company", AmountRUB: 1, StartDate: &start, EndDate: "2030-01-01"}))
		deposits, err := r.GetDeposits(ctx, 1)
		require.NoError(t, err)
		insurances, err := r.GetInsurances(ctx, 1)
		require.NoError(t, err)
		require.Equal(t, &start, deposits[0].StartDate)
		require.Equal(t, &start, insurances[0].StartDate)
		newStart := "2025-01-01"
		require.NoError(t, r.UpdateDeposit(ctx, 1, model.DepositUpdate{ID: deposits[0].ID, StartDate: &newStart}))
		require.NoError(t, r.UpdateInsurance(ctx, 1, model.InsuranceUpdate{ID: insurances[0].ID, StartDate: &newStart}))
		badEnd := "2024-01-01"
		require.ErrorIs(t, r.UpdateDeposit(ctx, 1, model.DepositUpdate{ID: deposits[0].ID, EndDate: &badEnd}), servererrors.ErrInvalidContract)
		require.ErrorIs(t, r.UpdateInsurance(ctx, 1, model.InsuranceUpdate{ID: insurances[0].ID, EndDate: &badEnd}), servererrors.ErrInvalidContract)
		badStart := "2031-01-01"
		require.ErrorIs(t, r.UpdateDeposit(ctx, 1, model.DepositUpdate{ID: deposits[0].ID, StartDate: &badStart}), servererrors.ErrInvalidContract)
		require.ErrorIs(t, r.UpdateInsurance(ctx, 1, model.InsuranceUpdate{ID: insurances[0].ID, StartDate: &badStart}), servererrors.ErrInvalidContract)
		deposits, err = r.GetDeposits(ctx, 1)
		require.NoError(t, err)
		insurances, err = r.GetInsurances(ctx, 1)
		require.NoError(t, err)
		require.Equal(t, &newStart, deposits[0].StartDate)
		require.Equal(t, &newStart, insurances[0].StartDate)
		require.NoError(t, r.RemoveDeposit(ctx, 1, deposits[0].ID))
		require.NoError(t, r.RemoveInsurance(ctx, 1, insurances[0].ID))
	})

	t.Run("resolved profile survives default switch and is user-bound", func(t *testing.T) {
		d := model.Deposit{Name: "scoped", BankName: "bank", AmountRUB: 100, AnnualInterest: 2, EndDate: "2030-01-01"}
		i := model.Insurance{Name: "scoped", CompanyName: "company", AmountRUB: 100, AnnualInterest: 2, EndDate: "2030-01-01"}
		c := model.CreditCard{Name: "scoped", BankName: "bank", MandatoryPaymentDate: "2030-01-01", MandatoryPaymentAmountRUB: 1, OverallLimitRUB: 100}
		require.NoError(t, r.AddDeposit(ctx, 1, d))
		require.NoError(t, r.AddInsurance(ctx, 1, i))
		require.NoError(t, r.AddCreditCard(ctx, 1, c))
		scoped, err := r.PersonalAssetProfileContext(ctx, 1)
		require.NoError(t, err)
		_, err = db.Exec(`UPDATE profiles SET is_default=(id=3) WHERE user_id=1`)
		require.NoError(t, err)
		deposits, err := r.GetDeposits(scoped, 1)
		require.NoError(t, err)
		require.Len(t, deposits, 1)
		cards, err := r.GetCreditCards(scoped, 1)
		require.NoError(t, err)
		require.Len(t, cards, 1)
		insurances, err := r.GetInsurances(scoped, 1)
		require.NoError(t, err)
		require.Len(t, insurances, 1)
		other, err := r.GetDeposits(scoped, 2)
		require.NoError(t, err)
		require.Empty(t, other)
		_, err = db.Exec(`UPDATE profiles SET is_default=(id=1) WHERE user_id=1`)
		require.NoError(t, err)
		require.NoError(t, r.RemoveDeposit(ctx, 1, deposits[0].ID))
		require.NoError(t, r.RemoveInsurance(ctx, 1, insurances[0].ID))
		require.NoError(t, r.RemoveCreditCard(ctx, 1, cards[0].ID))
	})

	t.Run("deposit", func(t *testing.T) {
		d := model.Deposit{Name: "old", BankName: "bank", AmountRUB: 100, AnnualInterest: 2, EndDate: "2030-01-01"}
		require.NoError(t, r.AddDeposit(ctx, 1, d))
		items, err := r.GetDeposits(ctx, 1)
		require.NoError(t, err)
		require.Len(t, items, 1)
		id := items[0].ID
		other, err := r.GetDeposits(ctx, 2)
		require.NoError(t, err)
		require.Empty(t, other)
		update := model.DepositUpdate{ID: id, Name: personalAssetPtr("new"), BankName: personalAssetPtr("new bank"), AnnualInterest: personalAssetPtr(5.0), EndDate: personalAssetPtr("2031-01-01")}
		require.ErrorIs(t, r.UpdateDeposit(ctx, 2, update), sql.ErrNoRows)
		require.ErrorIs(t, r.RemoveDeposit(ctx, 2, id), sql.ErrNoRows)
		require.ErrorIs(t, r.CloseDeposit(ctx, 2, model.DepositClose{ID: id, ClosingAmountRUB: 120}), sql.ErrNoRows)
		require.NoError(t, r.UpdateDeposit(ctx, 1, update))
		items, err = r.GetDeposits(ctx, 1)
		require.NoError(t, err)
		require.Equal(t, model.Deposit{ID: id, Name: "new", BankName: "new bank", AmountRUB: 100, AnnualInterest: 5, EndDate: "2031-01-01"}, items[0])
		require.NoError(t, r.UpdateDeposit(ctx, 1, model.DepositUpdate{ID: id, AnnualInterest: personalAssetPtr(0.0)}))
		items, err = r.GetDeposits(ctx, 1)
		require.NoError(t, err)
		require.Equal(t, model.Deposit{ID: id, Name: "new", BankName: "new bank", AmountRUB: 100, AnnualInterest: 0, EndDate: "2031-01-01"}, items[0])
		require.NoError(t, r.UpdateDeposit(ctx, 1, model.DepositUpdate{ID: id, Name: personalAssetPtr("partial")}))
		var snapshotName, snapshotBank string
		var snapshotAmount float64
		err = db.QueryRow(`SELECT name,bank_name,amount_rub FROM deposit_transactions WHERE profile_id=1 AND buying`).Scan(&snapshotName, &snapshotBank, &snapshotAmount)
		require.NoError(t, err)
		require.Equal(t, "partial", snapshotName)
		require.Equal(t, "new bank", snapshotBank)
		require.Equal(t, 100.0, snapshotAmount)
		require.Error(t, r.CloseDeposit(ctx, 1, model.DepositClose{ID: id, ClosingAmountRUB: -1}))
		items, err = r.GetDeposits(ctx, 1)
		require.NoError(t, err)
		require.Len(t, items, 1)
		require.NoError(t, r.CloseDeposit(ctx, 1, model.DepositClose{ID: id, ClosingAmountRUB: 120}))
		var count int
		var good bool
		err = db.QueryRow(`SELECT count(*), bool_and(name='partial' AND bank_name='new bank' AND created_at IS NOT NULL AND ((buying AND amount_rub=100) OR (NOT buying AND amount_rub=120))) FROM deposit_transactions WHERE profile_id=1`).Scan(&count, &good)
		require.NoError(t, err)
		require.Equal(t, 2, count)
		require.True(t, good)
		require.NoError(t, r.AddDeposit(ctx, 1, d))
		items, err = r.GetDeposits(ctx, 1)
		require.NoError(t, err)
		require.NoError(t, r.RemoveDeposit(ctx, 1, items[0].ID))
		d.EndDate = "invalid"
		require.Error(t, r.AddDeposit(ctx, 1, d))
		err = db.QueryRow(`SELECT count(*) FROM deposit_transactions`).Scan(&count)
		require.NoError(t, err)
		require.Equal(t, 2, count)
	})

	t.Run("insurance", func(t *testing.T) {
		d := model.Insurance{Name: "old", CompanyName: "company", AmountRUB: 200, AnnualInterest: 2, EndDate: "2030-01-01"}
		require.NoError(t, r.AddInsurance(ctx, 1, d))
		items, err := r.GetInsurances(ctx, 1)
		require.NoError(t, err)
		require.Len(t, items, 1)
		id := items[0].ID
		other, err := r.GetInsurances(ctx, 2)
		require.NoError(t, err)
		require.Empty(t, other)
		update := model.InsuranceUpdate{ID: id, Name: personalAssetPtr("new"), CompanyName: personalAssetPtr("new company"), AnnualInterest: personalAssetPtr(5.0), EndDate: personalAssetPtr("2031-01-01")}
		require.ErrorIs(t, r.UpdateInsurance(ctx, 2, update), sql.ErrNoRows)
		require.ErrorIs(t, r.RemoveInsurance(ctx, 2, id), sql.ErrNoRows)
		require.ErrorIs(t, r.CloseInsurance(ctx, 2, model.InsuranceClose{ID: id, ClosingAmountRUB: 240}), sql.ErrNoRows)
		require.NoError(t, r.UpdateInsurance(ctx, 1, update))
		items, err = r.GetInsurances(ctx, 1)
		require.NoError(t, err)
		require.Equal(t, model.Insurance{ID: id, Name: "new", CompanyName: "new company", AmountRUB: 200, AnnualInterest: 5, EndDate: "2031-01-01"}, items[0])
		require.NoError(t, r.UpdateInsurance(ctx, 1, model.InsuranceUpdate{ID: id, AnnualInterest: personalAssetPtr(0.0)}))
		items, err = r.GetInsurances(ctx, 1)
		require.NoError(t, err)
		require.Equal(t, model.Insurance{ID: id, Name: "new", CompanyName: "new company", AmountRUB: 200, AnnualInterest: 0, EndDate: "2031-01-01"}, items[0])
		require.NoError(t, r.UpdateInsurance(ctx, 1, model.InsuranceUpdate{ID: id, CompanyName: personalAssetPtr("partial company")}))
		var snapshotName, snapshotCompany string
		var snapshotAmount float64
		err = db.QueryRow(`SELECT name,company_name,amount_rub FROM insurance_transactions WHERE profile_id=1 AND buying`).Scan(&snapshotName, &snapshotCompany, &snapshotAmount)
		require.NoError(t, err)
		require.Equal(t, "new", snapshotName)
		require.Equal(t, "partial company", snapshotCompany)
		require.Equal(t, 200.0, snapshotAmount)
		require.Error(t, r.CloseInsurance(ctx, 1, model.InsuranceClose{ID: id, ClosingAmountRUB: -1}))
		items, err = r.GetInsurances(ctx, 1)
		require.NoError(t, err)
		require.Len(t, items, 1)
		require.NoError(t, r.CloseInsurance(ctx, 1, model.InsuranceClose{ID: id, ClosingAmountRUB: 240}))
		var count int
		var good bool
		err = db.QueryRow(`SELECT count(*), bool_and(name='new' AND company_name='partial company' AND created_at IS NOT NULL AND ((buying AND amount_rub=200) OR (NOT buying AND amount_rub=240))) FROM insurance_transactions WHERE profile_id=1`).Scan(&count, &good)
		require.NoError(t, err)
		require.Equal(t, 2, count)
		require.True(t, good)
		require.NoError(t, r.AddInsurance(ctx, 1, d))
		items, err = r.GetInsurances(ctx, 1)
		require.NoError(t, err)
		require.NoError(t, r.RemoveInsurance(ctx, 1, items[0].ID))
		d.EndDate = "invalid"
		require.Error(t, r.AddInsurance(ctx, 1, d))
		err = db.QueryRow(`SELECT count(*) FROM insurance_transactions`).Scan(&count)
		require.NoError(t, err)
		require.Equal(t, 2, count)
	})
	t.Run("down migration", func(t *testing.T) {
		down, err := os.ReadFile("../../../migrations/000012_add_contract_start_dates.down.sql")
		require.NoError(t, err)
		_, err = db.Exec(string(down))
		require.NoError(t, err)
		var count int
		err = db.QueryRow(`SELECT count(*) FROM information_schema.columns WHERE table_schema=current_schema() AND table_name IN ('deposits','insurances') AND column_name='start_date'`).Scan(&count)
		require.NoError(t, err)
		require.Zero(t, count)
	})
}
