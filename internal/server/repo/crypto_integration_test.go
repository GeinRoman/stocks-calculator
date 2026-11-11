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

// CRYPTO_TEST_DATABASE_URL must point to a disposable database with CREATE schema privileges.
func TestCryptoRepositoryIntegration(t *testing.T) {
	dsn := os.Getenv("CRYPTO_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set CRYPTO_TEST_DATABASE_URL to run PostgreSQL integration tests")
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
	schema := fmt.Sprintf("crypto_test_%x", random)
	_, err = admin.Exec("CREATE SCHEMA " + pq.QuoteIdentifier(schema))
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := admin.Exec("DROP SCHEMA " + pq.QuoteIdentifier(schema) + " CASCADE")
		require.NoError(t, err)
	})
	db, err := sql.Open("postgres", dsn+" search_path="+schema)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	_, err = db.Exec(`
		CREATE TABLE profiles (id SERIAL PRIMARY KEY, user_id INTEGER NOT NULL, is_default BOOLEAN NOT NULL);
		CREATE TABLE crypto_groups (id SERIAL PRIMARY KEY, profile_id INTEGER NOT NULL REFERENCES profiles(id) ON DELETE CASCADE, name VARCHAR(64) NOT NULL, weight INTEGER NOT NULL DEFAULT 0 CHECK (weight BETWEEN 0 AND 100), UNIQUE(profile_id,name), UNIQUE(id,profile_id));
		CREATE TABLE crypto_assets (id BIGSERIAL PRIMARY KEY, profile_id INTEGER NOT NULL REFERENCES profiles(id) ON DELETE CASCADE, group_id INTEGER NOT NULL, name TEXT NOT NULL, platform TEXT NOT NULL, url TEXT NOT NULL DEFAULT '', annual_interest DOUBLE PRECISION NOT NULL CHECK (annual_interest >= 0), amount NUMERIC NOT NULL CHECK (amount > 0), UNIQUE(profile_id,name), FOREIGN KEY(group_id,profile_id) REFERENCES crypto_groups(id,profile_id));
		CREATE TABLE crypto_transactions (id BIGSERIAL PRIMARY KEY, profile_id INTEGER NOT NULL REFERENCES profiles(id) ON DELETE CASCADE, asset_id BIGINT NOT NULL, asset_name TEXT NOT NULL, platform TEXT NOT NULL, amount NUMERIC NOT NULL CHECK (amount > 0), buying BOOLEAN NOT NULL, price DOUBLE PRECISION NOT NULL CHECK (price > 0), currency VARCHAR(16) NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP);
		INSERT INTO profiles (user_id,is_default) VALUES (1,true),(2,true),(1,false);
	`)
	require.NoError(t, err)
	ctx := context.Background()
	r := New(db)
	for _, user := range []int{1, 2} {
		require.NoError(t, r.InsertCryptoGroups(ctx, user, []model.CryptoGroup{{Name: "old"}, {Name: "new"}}))
	}
	op := model.CryptoOperation{CryptoAsset: model.CryptoAsset{Name: "BTC", GroupName: "old", Platform: "A", URL: "old", AnnualInterest: 1, Amount: 1.5}, Price: model.Money{Price: 10, Currency: "USD"}}
	require.NoError(t, r.TransactCrypto(ctx, 1, []model.CryptoOperation{op}, true))
	require.NoError(t, r.TransactCrypto(ctx, 2, []model.CryptoOperation{op}, true))
	assets, err := r.GetCryptoAssets(ctx, 1)
	require.NoError(t, err)
	require.Len(t, assets, 1)
	id := assets[0].ID
	require.Positive(t, id)
	other := op
	other.Name = "ETH"
	require.NoError(t, r.TransactCrypto(ctx, 1, []model.CryptoOperation{other}, true))
	update := model.CryptoAssetUpdate{ID: id, Name: "ETH", Platform: "B", URL: "new", AnnualInterest: 3}
	require.ErrorIs(t, r.UpdateCryptoAsset(ctx, 1, update), servererrors.ErrCryptoExists)
	update.Name = "renamed"
	require.ErrorIs(t, r.UpdateCryptoAsset(ctx, 2, update), servererrors.ErrCryptoNotFound)
	require.NoError(t, r.UpdateCryptoAsset(ctx, 1, update))
	assets, err = r.GetCryptoAssets(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, model.CryptoAsset{ID: id, Name: "renamed", GroupName: "old", Platform: "B", URL: "new", AnnualInterest: 3, Amount: 1.5}, assets[1])
	require.ErrorIs(t, r.RemoveCryptoGroups(ctx, 1, []model.CryptoGroup{{Name: "old"}}), servererrors.ErrCryptoGroupNotEmpty)
	byID := model.CryptoOperation{CryptoAsset: model.CryptoAsset{ID: id, Amount: 0.5, Name: "untrusted", Platform: "untrusted"}, Price: model.Money{Price: 20, Currency: "EUR"}}
	require.ErrorIs(t, r.TransactCrypto(ctx, 2, []model.CryptoOperation{byID}, false), servererrors.ErrCryptoNotFound)
	require.NoError(t, r.TransactCrypto(ctx, 1, []model.CryptoOperation{byID}, false))
	missing := byID
	missing.ID = -1
	require.ErrorIs(t, r.TransactCrypto(ctx, 1, []model.CryptoOperation{byID, missing}, false), servererrors.ErrCryptoNotFound)
	byID.Amount = 1
	require.NoError(t, r.TransactCrypto(ctx, 1, []model.CryptoOperation{byID}, false))
	rows, err := db.Query(`SELECT asset_id,asset_name,platform,amount,buying,created_at IS NOT NULL FROM crypto_transactions WHERE profile_id=1 ORDER BY id`)
	require.NoError(t, err)
	defer rows.Close()
	var ids []int64
	var names, platforms []string
	var amounts []float64
	for rows.Next() {
		var txID int64
		var name, platform string
		var amount float64
		var buying, hasTime bool
		require.NoError(t, rows.Scan(&txID, &name, &platform, &amount, &buying, &hasTime))
		require.True(t, hasTime)
		if txID == id {
			ids = append(ids, txID)
			names = append(names, name)
			platforms = append(platforms, platform)
			amounts = append(amounts, amount)
		}
	}
	require.NoError(t, rows.Err())
	require.Equal(t, []int64{id, id, id}, ids)
	require.Equal(t, []string{"BTC", "renamed", "renamed"}, names)
	require.Equal(t, []string{"A", "B", "B"}, platforms)
	require.Equal(t, []float64{1.5, 0.5, 1}, amounts)
	require.NoError(t, rows.Close())

	large := op
	large.Name = "large"
	large.Amount = 1e16
	require.NoError(t, r.TransactCrypto(ctx, 1, []model.CryptoOperation{large}, true))
	var largeID int64
	require.NoError(t, db.QueryRow(`SELECT id FROM crypto_assets WHERE profile_id=1 AND name='large'`).Scan(&largeID))
	change := model.CryptoOperation{CryptoAsset: model.CryptoAsset{ID: largeID, Amount: 3}, Price: op.Price}
	require.ErrorIs(t, r.TransactCrypto(ctx, 1, []model.CryptoOperation{change}, true), servererrors.ErrInvalidCrypto)
	require.ErrorIs(t, r.TransactCrypto(ctx, 1, []model.CryptoOperation{change}, false), servererrors.ErrInvalidCrypto)
	var balance float64
	var count int
	require.NoError(t, db.QueryRow(`SELECT amount FROM crypto_assets WHERE id=$1`, largeID).Scan(&balance))
	require.Equal(t, 1e16, balance)
	require.NoError(t, db.QueryRow(`SELECT count(*) FROM crypto_transactions WHERE asset_id=$1`, largeID).Scan(&count))
	require.Equal(t, 1, count)
	change.Amount = 4
	require.NoError(t, r.TransactCrypto(ctx, 1, []model.CryptoOperation{change}, true))
	require.NoError(t, db.QueryRow(`SELECT amount FROM crypto_assets WHERE id=$1`, largeID).Scan(&balance))
	require.Equal(t, 1e16+4, balance)
	require.NoError(t, r.TransactCrypto(ctx, 1, []model.CryptoOperation{change}, false))
	require.NoError(t, db.QueryRow(`SELECT amount FROM crypto_assets WHERE id=$1`, largeID).Scan(&balance))
	require.Equal(t, 1e16, balance)

	decimal := op
	decimal.Name, decimal.Amount = "decimal", 0.1
	require.NoError(t, r.TransactCrypto(ctx, 1, []model.CryptoOperation{decimal}, true))
	var decimalID int64
	require.NoError(t, db.QueryRow(`SELECT id FROM crypto_assets WHERE profile_id=1 AND name='decimal'`).Scan(&decimalID))
	delta := model.CryptoOperation{CryptoAsset: model.CryptoAsset{ID: decimalID, Amount: 0.2}, Price: op.Price}
	require.NoError(t, r.TransactCrypto(ctx, 1, []model.CryptoOperation{delta}, true))
	var stored string
	require.NoError(t, db.QueryRow(`SELECT amount::text FROM crypto_assets WHERE id=$1`, decimalID).Scan(&stored))
	require.Equal(t, "0.3", stored)
	delta.Amount = 0.1
	require.NoError(t, r.TransactCrypto(ctx, 1, []model.CryptoOperation{delta}, false))
	delta.Amount = 0
	require.NoError(t, r.TransactCrypto(ctx, 1, []model.CryptoOperation{delta}, false))
	var net string
	require.NoError(t, db.QueryRow(`SELECT sum(CASE WHEN buying THEN amount ELSE -amount END)::text FROM crypto_transactions WHERE asset_id=$1`, decimalID).Scan(&net))
	require.Equal(t, "0.0", net)
}
