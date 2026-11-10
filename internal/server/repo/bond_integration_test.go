package repo

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/servererrors"
)

// Set BOND_TEST_DATABASE_URL to a disposable PostgreSQL database with CREATE
// schema privileges. The test creates and drops only its own isolated schema.
func TestBondRepositoryIntegration(t *testing.T) {
	dsn := os.Getenv("BOND_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set BOND_TEST_DATABASE_URL to run PostgreSQL integration tests")
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
	schema := fmt.Sprintf("bond_test_%x", random)
	_, err = admin.Exec("CREATE SCHEMA " + pq.QuoteIdentifier(schema))
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := admin.Exec("DROP SCHEMA " + pq.QuoteIdentifier(schema) + " CASCADE")
		if err != nil {
			t.Errorf("drop test schema: %v", err)
		}
	})

	db, err := sql.Open("postgres", dsn+" search_path="+schema)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	migrations, err := filepath.Glob("../../../migrations/*.up.sql")
	require.NoError(t, err)
	require.NotEmpty(t, migrations)
	for _, path := range migrations {
		migration, err := os.ReadFile(path)
		require.NoError(t, err)
		var legacyUser, legacyProfile int
		if filepath.Base(path) == "000009_transaction_currency.up.sql" {
			require.NoError(t, db.QueryRow(`INSERT INTO users (name, password) VALUES ('legacy', 'unused') RETURNING id`).Scan(&legacyUser))
			require.NoError(t, db.QueryRow(`INSERT INTO profiles (user_id, name, is_default) VALUES ($1, 'legacy', true) RETURNING id`, legacyUser).Scan(&legacyProfile))
			_, err = db.Exec(`INSERT INTO transactions (profile_id, instrument_code, instrument_type, lot_amount, buying, lot_size, price_in_rub, datetime) VALUES ($1, 'LEGACY', 'stock', 1, true, 1, 42.5, NOW())`, legacyProfile)
			require.NoError(t, err)
		}
		_, err = db.Exec(string(migration))
		require.NoError(t, err, "migration %s", path)
		if legacyUser != 0 {
			var currency string
			var price float64
			require.NoError(t, db.QueryRow(`SELECT currency, price FROM transactions WHERE profile_id = $1`, legacyProfile).Scan(&currency, &price))
			require.Equal(t, "RUB", currency)
			require.Equal(t, 42.5, price)
			_, err = db.Exec(`DELETE FROM users WHERE id = $1`, legacyUser)
			require.NoError(t, err)
		}
	}

	ctx := context.Background()
	r := New(db)
	var user, otherUser, profile, otherProfile int
	require.NoError(t, db.QueryRow(`INSERT INTO users (name, password) VALUES ('one', 'unused') RETURNING id`).Scan(&user))
	require.NoError(t, db.QueryRow(`INSERT INTO users (name, password) VALUES ('two', 'unused') RETURNING id`).Scan(&otherUser))
	require.NoError(t, db.QueryRow(`INSERT INTO profiles (user_id, name, is_default) VALUES ($1, 'default', true) RETURNING id`, user).Scan(&profile))
	require.NoError(t, db.QueryRow(`INSERT INTO profiles (user_id, name, is_default) VALUES ($1, 'default', true) RETURNING id`, otherUser).Scan(&otherProfile))
	require.NoError(t, r.InsertProfile(ctx, "secondary", user, false))

	group := model.BondGroup{Name: "shared"}
	require.NoError(t, r.InsertBondGroups(ctx, user, []model.BondGroup{group}))
	require.NoError(t, r.InsertBondGroups(ctx, otherUser, []model.BondGroup{group}))
	require.NoError(t, r.UpdateBondWeights(ctx, user, []model.BondGroup{{Name: group.Name, Weight: 100}}))
	groups, err := r.GetBondGroups(ctx, otherUser)
	require.NoError(t, err)
	require.Equal(t, []model.BondGroup{group}, groups)

	bond := model.BondOperation{Name: "Bond", Code: "BOND", GroupName: group.Name, LotAmount: 5, LotSize: 10, Price: model.Money{Currency: "RUB", Price: 98.5}}
	otherBond := bond
	otherBond.LotAmount = 7
	require.NoError(t, r.InsertBonds(ctx, user, []model.BondOperation{bond}))
	require.NoError(t, r.InsertBonds(ctx, otherUser, []model.BondOperation{otherBond}))

	// A later failure must roll back both the holding and its transaction.
	missing := model.BondOperation{Name: "Missing", Code: "MISSING", GroupName: "missing", LotAmount: 1, LotSize: 1}
	first := model.BondOperation{Name: "First", Code: "FIRST", GroupName: group.Name, LotAmount: 1, LotSize: 1}
	require.ErrorIs(t, r.InsertBonds(ctx, user, []model.BondOperation{first, missing}), servererrors.ErrGroupNotFound)
	var count int
	require.NoError(t, db.QueryRow(`SELECT count(*) FROM transactions WHERE instrument_code = 'FIRST'`).Scan(&count))
	require.Zero(t, count)

	bond.LotAmount = 3
	bond.Price = model.Money{Currency: "USD", Price: 101.25}
	require.NoError(t, r.UpdateBondsAmount(ctx, user, []model.BondOperation{bond}, []int{-2}))
	stored, err := r.GetBonds(ctx, user)
	require.NoError(t, err)
	require.Equal(t, []model.Bond{{Name: bond.Name, Code: bond.Code, GroupName: bond.GroupName, LotAmount: 3, LotSize: 10}}, stored)
	otherStored, err := r.GetBonds(ctx, otherUser)
	require.NoError(t, err)
	require.Equal(t, []model.Bond{{Name: otherBond.Name, Code: otherBond.Code, GroupName: otherBond.GroupName, LotAmount: 7, LotSize: 10}}, otherStored)

	require.NoError(t, r.SetDefaultProfile(ctx, "secondary", user))
	stored, err = r.GetBonds(ctx, user)
	require.NoError(t, err)
	require.Empty(t, stored)
	require.Error(t, r.RemoveBonds(ctx, user, []model.BondOperation{bond}))
	require.NoError(t, r.SetDefaultProfile(ctx, "default", user))
	require.NoError(t, r.RemoveBonds(ctx, user, []model.BondOperation{bond}))
	stored, err = r.GetBonds(ctx, user)
	require.NoError(t, err)
	require.Empty(t, stored)
	otherStored, err = r.GetBonds(ctx, otherUser)
	require.NoError(t, err)
	require.Len(t, otherStored, 1)
	require.Equal(t, 7, otherStored[0].LotAmount)

	rows, err := db.Query(`SELECT profile_id, instrument_code, instrument_type, lot_amount, buying, lot_size, price, currency FROM transactions ORDER BY id`)
	require.NoError(t, err)
	defer rows.Close()
	type transaction struct {
		profile, amount, size int
		code, kind, currency  string
		buying                bool
		price                 float64
	}
	var got []transaction
	for rows.Next() {
		var tx transaction
		require.NoError(t, rows.Scan(&tx.profile, &tx.code, &tx.kind, &tx.amount, &tx.buying, &tx.size, &tx.price, &tx.currency))
		got = append(got, tx)
	}
	require.NoError(t, rows.Err())
	require.Equal(t, []transaction{
		{profile, 5, 10, "BOND", "bond", "RUB", true, 98.5},
		{otherProfile, 7, 10, "BOND", "bond", "RUB", true, 98.5},
		{profile, 2, 10, "BOND", "bond", "USD", false, 101.25},
		{profile, 3, 10, "BOND", "bond", "USD", false, 101.25},
	}, got)
	require.NoError(t, rows.Close())
	require.NoError(t, r.InsertGroups(ctx, user, []model.StockGroup{{Name: "stocks"}}))
	require.NoError(t, r.InsertStocks(ctx, user, []model.Stock{{Name: "Stock", Code: "STOCK", GroupName: "stocks", LotAmount: 1, LotSize: 1, Price: 123.45}}))
	var stockCurrency string
	var stockPrice float64
	require.NoError(t, db.QueryRow(`SELECT currency, price FROM transactions WHERE instrument_code = 'STOCK'`).Scan(&stockCurrency, &stockPrice))
	require.Equal(t, "RUB", stockCurrency)
	require.Equal(t, 123.45, stockPrice)
	down, err := os.ReadFile("../../../migrations/000009_transaction_currency.down.sql")
	require.NoError(t, err)
	_, err = db.Exec(string(down))
	require.Error(t, err, "rollback must not relabel foreign currency transactions as RUB")
	require.NoError(t, db.QueryRow(`SELECT count(*) FROM transactions WHERE currency = 'USD'`).Scan(&count))
	require.Equal(t, 2, count)
	_, err = db.Exec(`DELETE FROM transactions WHERE currency <> 'RUB'`)
	require.NoError(t, err)
	_, err = db.Exec(string(down))
	require.NoError(t, err)
	require.NoError(t, db.QueryRow(`SELECT price_in_rub FROM transactions WHERE instrument_code = 'STOCK'`).Scan(&stockPrice))
	require.Equal(t, 123.45, stockPrice)
}
