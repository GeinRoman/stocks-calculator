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
	"stocks_calculator/internal/server/app"
)

// SNAPSHOT_TEST_DATABASE_URL must point to a disposable PostgreSQL database.
func TestSnapshotIntegration(t *testing.T) {
	dsn := os.Getenv("SNAPSHOT_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set SNAPSHOT_TEST_DATABASE_URL to run PostgreSQL integration tests")
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
	schema := fmt.Sprintf("snapshot_test_%x", random)
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
		b, err := os.ReadFile(path)
		require.NoError(t, err)
		_, err = db.Exec(string(b))
		require.NoError(t, err, path)
	}
	_, err = db.Exec(`
		INSERT INTO users(id,name,password) VALUES (1,'owner','unused'),(2,'other','unused');
		INSERT INTO profiles(id,user_id,name,is_default) VALUES (1,1,'main',true),(2,1,'secondary',false),(3,2,'untouched',true);
		INSERT INTO groups(id,profile_id,name,weight) VALUES (1,1,'stocks',100);
		INSERT INTO stocks(group_id,name,code,lot_amount,lot_size) VALUES (1,'stock','S',2,0);
		INSERT INTO bond_groups(id,profile_id,name,weight) VALUES (1,1,'bonds',100);
		INSERT INTO bonds(group_id,name,code,lot_amount,lot_size) VALUES (1,'bond','B',1,1);
		INSERT INTO crypto_groups(id,profile_id,name,weight) VALUES (1,1,'crypto',100);
		INSERT INTO crypto_assets(id,profile_id,group_id,name,platform,url,annual_interest,amount) VALUES (100,1,1,'coin','exchange','',2,0.1234567890123456789);
		INSERT INTO crypto_transactions(profile_id,asset_id,asset_name,platform,amount,buying,price,currency,created_at) VALUES (1,100,'coin','exchange',0.1234567890123456789,true,10,'USD','2020-01-01Z'),(1,200,'sold','exchange',1,false,5,'RUB','2020-01-01Z');
		INSERT INTO transactions(profile_id,instrument_code,instrument_type,lot_amount,buying,lot_size,price,currency,datetime) VALUES (1,'S','stock',2,true,0,5,'RUB','2020-01-01Z'),(1,'B','bond',1,true,1,10,'USD','2020-01-01Z');
		INSERT INTO deposit_transactions(id,profile_id,name,bank_name,amount_rub,buying,created_at) VALUES (100,1,'deposit','bank',100,true,'2020-01-01Z');
		INSERT INTO deposits(profile_id,acquisition_transaction_id,name,bank_name,amount_rub,annual_interest,start_date,end_date) VALUES (1,100,'deposit','bank',100,5,'2020-01-01','2030-01-01');
		INSERT INTO insurance_transactions(id,profile_id,name,company_name,amount_rub,buying,created_at) VALUES (100,1,'insurance','company',200,true,'2020-01-01Z');
		INSERT INTO insurances(profile_id,acquisition_transaction_id,name,company_name,amount_rub,annual_interest,start_date,end_date) VALUES (1,100,'insurance','company',200,5,'2020-01-01','2030-01-01');
		INSERT INTO credit_cards(profile_id,name,bank_name,mandatory_payment_date,mandatory_payment_amount_rub,overall_limit_rub) VALUES (2,'card','bank','2030-01-01',1,100);`)
	require.NoError(t, err)
	r := New(db)
	svc := app.New(r, nil, nil)
	ctx := context.Background()
	before, err := svc.GetSnapshot(ctx, 1)
	require.NoError(t, err)
	require.Len(t, before.Profiles, 2)
	require.Equal(t, "0.1234567890123456789", before.Profiles[0].CryptoGroups[0].Assets[0].Amount)
	require.NoError(t, svc.InstallSnapshot(ctx, 1, before))
	after, err := svc.GetSnapshot(ctx, 1)
	require.NoError(t, err)
	require.Len(t, after.Profiles, 2)
	require.Equal(t, before.Profiles[0].Name, after.Profiles[0].Name)
	require.Equal(t, before.Profiles[0].StockGroups, after.Profiles[0].StockGroups)
	require.Equal(t, before.Profiles[0].BondGroups, after.Profiles[0].BondGroups)
	require.Equal(t, before.Profiles[0].CryptoGroups[0].Assets[0].Amount, after.Profiles[0].CryptoGroups[0].Assets[0].Amount)
	require.Equal(t, after.Profiles[0].CryptoGroups[0].Assets[0].ID, after.Profiles[0].CryptoTransactions[0].AssetID)
	require.NotEqual(t, after.Profiles[0].CryptoTransactions[0].AssetID, after.Profiles[0].CryptoTransactions[1].AssetID)
	require.Equal(t, after.Profiles[0].DepositTransactions[0].ID, after.Profiles[0].Deposits[0].AcquisitionTransactionID)
	require.Equal(t, after.Profiles[0].InsuranceTransactions[0].ID, after.Profiles[0].Insurances[0].AcquisitionTransactionID)
	require.Len(t, after.Profiles[0].Transactions, 2)
	require.Len(t, after.Profiles[1].CreditCards, 1)
	other, err := svc.GetSnapshot(ctx, 2)
	require.NoError(t, err)
	require.Equal(t, "untouched", other.Profiles[0].Name)
	bad := after
	duplicate := bad.Profiles[0]
	duplicate.Default = false
	bad.Profiles = append(bad.Profiles, duplicate)
	require.Error(t, svc.InstallSnapshot(ctx, 1, bad))
	unchanged, err := svc.GetSnapshot(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, after, unchanged)
}
