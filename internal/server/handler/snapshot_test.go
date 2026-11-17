package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/servererrors"
)

type snapshotTestApp struct {
	App
	user     int
	snapshot model.Snapshot
	err      error
	installs int
}

func (a *snapshotTestApp) GetSnapshot(_ context.Context, user int) (model.Snapshot, error) {
	a.user = user
	return a.snapshot, a.err
}
func (a *snapshotTestApp) InstallSnapshot(_ context.Context, user int, s model.Snapshot) error {
	a.user = user
	a.snapshot = s
	a.installs++
	return a.err
}

func TestSnapshotRoutes(t *testing.T) {
	a := &snapshotTestApp{snapshot: model.Snapshot{Version: 1, Profiles: []model.SnapshotProfile{}}}
	h := NewHttpServer(a, bondTestTokens{}, 0).Handler
	for _, path := range []string{"/getsnapshot", "/installsnapshot"} {
		method := http.MethodGet
		if path == "/installsnapshot" {
			method = http.MethodPost
		}
		require.Equal(t, http.StatusUnauthorized, bondRequest(h, method, path, "", "").Code)
	}
	w := bondRequest(h, http.MethodGet, "/getsnapshot", "", "alice")
	require.Equal(t, 200, w.Code)
	require.Equal(t, 11, a.user)
	require.Equal(t, `attachment; filename="snapshot.json"`, w.Header().Get("Content-Disposition"))
	require.Equal(t, "no-store", w.Header().Get("Cache-Control"))
	require.JSONEq(t, `{"version":1,"profiles":[]}`, w.Body.String())
	w = bondRequest(h, http.MethodPost, "/installsnapshot", w.Body.String(), "bob")
	require.Equal(t, 200, w.Code)
	require.Equal(t, 22, a.user)
	require.Equal(t, 1, a.installs)
	for _, invalid := range []string{`{}`, `{"version":1}`, `{"version":1,"profiles":null}`, `{"version":1,"profiles":[],"unknown":1}`, `{"version":1,"profiles":[]} {}`} {
		w = bondRequest(h, http.MethodPost, "/installsnapshot", invalid, "alice")
		require.Equal(t, 400, w.Code, invalid)
		require.Equal(t, 1, a.installs)
	}
	p := model.SnapshotProfile{Name: "test", StockGroups: []model.SnapshotStockGroup{}, BondGroups: []model.SnapshotBondGroup{}, CryptoGroups: []model.SnapshotCryptoGroup{}, Deposits: []model.SnapshotDeposit{}, Insurances: []model.SnapshotInsurance{}, CreditCards: []model.CreditCard{}, Transactions: []model.SnapshotTransaction{}, CryptoTransactions: []model.SnapshotCryptoTransaction{}, DepositTransactions: []model.DepositTransaction{}, InsuranceTransactions: []model.InsuranceTransaction{}}
	data, err := json.Marshal(model.Snapshot{Version: 1, Profiles: []model.SnapshotProfile{p}})
	require.NoError(t, err)
	w = bondRequest(h, http.MethodPost, "/installsnapshot", string(data), "alice")
	require.Equal(t, 200, w.Code)
	broken := strings.Replace(string(data), `"default":false,`, "", 1)
	require.Equal(t, 400, bondRequest(h, http.MethodPost, "/installsnapshot", broken, "alice").Code)
	a.err = servererrors.ErrInvalidSnapshot
	require.Equal(t, 400, bondRequest(h, http.MethodPost, "/installsnapshot", string(data), "alice").Code)
	a.err = errors.New("failure")
	require.Equal(t, 500, bondRequest(h, http.MethodPost, "/installsnapshot", string(data), "alice").Code)
}
