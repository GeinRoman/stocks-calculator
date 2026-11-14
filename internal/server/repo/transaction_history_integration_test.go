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
	"time"

	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/servererrors"
)

// TRANSACTION_TEST_DATABASE_URL must refer to a disposable PostgreSQL database.
func TestTransactionHistoryIntegration(t *testing.T) {
	dsn := os.Getenv("TRANSACTION_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TRANSACTION_TEST_DATABASE_URL to run PostgreSQL integration tests")
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
	schema := fmt.Sprintf("transaction_history_%x", random)
	_, err = admin.Exec("CREATE SCHEMA " + pq.QuoteIdentifier(schema))
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := admin.Exec("DROP SCHEMA " + pq.QuoteIdentifier(schema) + " CASCADE")
		require.NoError(t, err)
	})
	db, err := sql.Open("postgres", dsn+" search_path="+schema)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	paths, err := filepath.Glob("../../../migrations/*.up.sql")
	require.NoError(t, err)
	for _, path := range paths {
		contents, err := os.ReadFile(path)
		require.NoError(t, err)
		_, err = db.Exec(string(contents))
		require.NoError(t, err, path)
	}
	_, err = db.Exec(`
		INSERT INTO users (id,name,password) VALUES (1,'history1','unused'),(2,'history2','unused');
		INSERT INTO profiles (id,user_id,name,is_default) VALUES (1,1,'default',true),(2,2,'default',true),(3,1,'other',false);
		INSERT INTO deposit_transactions(profile_id,name,bank_name,amount_rub,buying,created_at)
		 SELECT 1,'deposit','bank',1,true,'2026-01-01T00:00:00Z' FROM generate_series(1,105);
		INSERT INTO insurance_transactions(profile_id,name,company_name,amount_rub,buying,created_at)
		 SELECT 1,'insurance','company',1,true,'2026-01-01T00:00:00Z' FROM generate_series(1,105);
		INSERT INTO crypto_transactions(profile_id,asset_id,asset_name,platform,amount,buying,price,currency,created_at)
		 SELECT 1,1,'crypto','platform',1,true,1,'RUB','2026-01-01T00:00:00Z' FROM generate_series(1,105);
		INSERT INTO transactions(profile_id,instrument_code,instrument_type,lot_amount,buying,lot_size,price,currency,datetime)
		 SELECT 1,'stock','stock',1,true,1,1,'RUB','2026-01-01T00:00:00Z' FROM generate_series(1,105);
		INSERT INTO transactions(profile_id,instrument_code,instrument_type,lot_amount,buying,lot_size,price,currency,datetime)
		 VALUES (1,'bond','bond',1,true,1,1,'RUB','2026-01-02T00:00:00Z');
		INSERT INTO deposit_transactions(profile_id,name,bank_name,amount_rub,buying,created_at)
		 VALUES (2,'other','bank',1,true,'2026-01-02T00:00:00Z'),(3,'other profile','bank',1,true,'2026-01-02T00:00:00Z');`)
	require.NoError(t, err)
	r := New(db)
	ctx := context.Background()
	type entry struct {
		id int64
		at time.Time
	}
	getters := map[string]func(int, model.TransactionCursor) ([]entry, error){
		"deposit": func(u int, c model.TransactionCursor) ([]entry, error) {
			v, e := r.GetDepositTransactions(ctx, u, c)
			out := []entry{}
			for _, x := range v {
				out = append(out, entry{x.ID, x.CreatedAt})
			}
			return out, e
		},
		"insurance": func(u int, c model.TransactionCursor) ([]entry, error) {
			v, e := r.GetInsuranceTransactions(ctx, u, c)
			out := []entry{}
			for _, x := range v {
				out = append(out, entry{x.ID, x.CreatedAt})
			}
			return out, e
		},
		"crypto": func(u int, c model.TransactionCursor) ([]entry, error) {
			v, e := r.GetCryptoTransactions(ctx, u, c)
			out := []entry{}
			for _, x := range v {
				out = append(out, entry{x.ID, x.CreatedAt})
			}
			return out, e
		},
		"stock": func(u int, c model.TransactionCursor) ([]entry, error) {
			v, e := r.GetStockTransactions(ctx, u, c)
			out := []entry{}
			for _, x := range v {
				out = append(out, entry{x.ID, x.Time})
			}
			return out, e
		},
	}
	for name, get := range getters {
		t.Run(name, func(t *testing.T) {
			first, err := get(1, model.TransactionCursor{})
			require.NoError(t, err)
			require.Len(t, first, 100)
			for i := 1; i < len(first); i++ {
				require.Greater(t, first[i-1].id, first[i].id)
			}
			last := first[len(first)-1]
			second, err := get(1, model.TransactionCursor{Before: &last.at, BeforeID: &last.id})
			require.NoError(t, err)
			require.Len(t, second, 5)
			require.Less(t, second[0].id, last.id)
			exclusive, err := get(1, model.TransactionCursor{Before: &last.at})
			require.NoError(t, err)
			require.Empty(t, exclusive)
			_, err = get(999, model.TransactionCursor{})
			require.ErrorIs(t, err, servererrors.ErrNoDefaultProfile)
			other, err := get(2, model.TransactionCursor{})
			require.NoError(t, err)
			if name == "deposit" {
				require.Len(t, other, 1)
			} else {
				require.Empty(t, other)
			}
		})
	}
}
