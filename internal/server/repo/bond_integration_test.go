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
		_, err = db.Exec(string(migration))
		require.NoError(t, err, "migration %s", path)
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

	bond := model.Bond{Name: "Bond", Code: "BOND", GroupName: group.Name, LotAmount: 5, LotSize: 10, CleanPrice: 98.5}
	otherBond := bond
	otherBond.LotAmount = 7
	require.NoError(t, r.InsertBonds(ctx, user, []model.Bond{bond}))
	require.NoError(t, r.InsertBonds(ctx, otherUser, []model.Bond{otherBond}))

	// A later failure must roll back both the holding and its transaction.
	missing := model.Bond{Name: "Missing", Code: "MISSING", GroupName: "missing", LotAmount: 1, LotSize: 1}
	first := model.Bond{Name: "First", Code: "FIRST", GroupName: group.Name, LotAmount: 1, LotSize: 1}
	require.ErrorIs(t, r.InsertBonds(ctx, user, []model.Bond{first, missing}), servererrors.ErrGroupNotFound)
	var count int
	require.NoError(t, db.QueryRow(`SELECT count(*) FROM transactions WHERE instrument_code = 'FIRST'`).Scan(&count))
	require.Zero(t, count)

	bond.LotAmount = 3
	bond.CleanPrice = 101.25
	require.NoError(t, r.UpdateBondsAmount(ctx, user, []model.Bond{bond}, []int{-2}))
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
	require.Error(t, r.RemoveBonds(ctx, user, []model.Bond{bond}))
	require.NoError(t, r.SetDefaultProfile(ctx, "default", user))
	require.NoError(t, r.RemoveBonds(ctx, user, []model.Bond{bond}))
	stored, err = r.GetBonds(ctx, user)
	require.NoError(t, err)
	require.Empty(t, stored)
	otherStored, err = r.GetBonds(ctx, otherUser)
	require.NoError(t, err)
	require.Len(t, otherStored, 1)
	require.Equal(t, 7, otherStored[0].LotAmount)

	rows, err := db.Query(`SELECT profile_id, instrument_code, instrument_type, lot_amount, buying, lot_size, price_in_rub FROM transactions ORDER BY id`)
	require.NoError(t, err)
	defer rows.Close()
	type transaction struct {
		profile, amount, size int
		code, kind            string
		buying                bool
		price                 float64
	}
	var got []transaction
	for rows.Next() {
		var tx transaction
		require.NoError(t, rows.Scan(&tx.profile, &tx.code, &tx.kind, &tx.amount, &tx.buying, &tx.size, &tx.price))
		got = append(got, tx)
	}
	require.NoError(t, rows.Err())
	require.Equal(t, []transaction{
		{profile, 5, 10, "BOND", "bond", true, 98.5},
		{otherProfile, 7, 10, "BOND", "bond", true, 98.5},
		{profile, 2, 10, "BOND", "bond", false, 101.25},
		{profile, 3, 10, "BOND", "bond", false, 101.25},
	}, got)
}
