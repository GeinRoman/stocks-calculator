package handler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/servererrors"
)

type timeoutBondApp struct {
	App
	err error
	ctx context.Context
}

func (a *timeoutBondApp) FindBonds(ctx context.Context, _ int, _ []string) ([]model.FoundBond, error) {
	a.ctx = ctx
	if a.err != nil {
		return nil, a.err
	}
	return nil, ctx.Err()
}

func TestFindBondsTimeoutAndCancellation(t *testing.T) {
	expired, cancelExpired := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancelExpired()
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	for _, tc := range []struct {
		name   string
		ctx    context.Context
		err    error
		status int
	}{
		{"expired request", expired, nil, http.StatusGatewayTimeout},
		{"canceled request", canceled, nil, http.StatusRequestTimeout},
		{"wrapped upstream deadline", context.Background(), fmt.Errorf("%w: %w", servererrors.ErrMoexUnhandled, context.DeadlineExceeded), http.StatusGatewayTimeout},
		{"wrapped upstream cancellation", context.Background(), fmt.Errorf("%w: %w", servererrors.ErrMoexUnhandled, context.Canceled), http.StatusRequestTimeout},
	} {
		t.Run(tc.name, func(t *testing.T) {
			app := &timeoutBondApp{err: tc.err}
			server := NewHttpServer(app, bondTestTokens{}, 0)
			r := httptest.NewRequest(http.MethodPost, "/findbonds", strings.NewReader(`["SU1"]`)).WithContext(tc.ctx)
			r.Header.Set("Authorization", "Bearer alice")
			w := httptest.NewRecorder()
			server.Handler.ServeHTTP(w, r)
			if w.Code != tc.status {
				t.Fatalf("status = %d, want %d", w.Code, tc.status)
			}
			if app.ctx == nil || app.ctx.Err() == nil {
				t.Fatal("search context was not canceled")
			}
		})
	}
}

func TestBondSearchAndServerTimeouts(t *testing.T) {
	app := &timeoutBondApp{err: servererrors.ErrMoexUnhandled}
	server := NewHttpServer(app, bondTestTokens{}, 0)
	before := time.Now()
	w := bondRequest(server.Handler, http.MethodPost, "/findbonds", `["SU1"]`, "alice")
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", w.Code)
	}
	deadline, ok := app.ctx.Deadline()
	if !ok || deadline.Before(before.Add(bondSearchTimeout)) || deadline.After(time.Now().Add(bondSearchTimeout)) {
		t.Fatalf("unexpected search deadline %v (set = %v)", deadline, ok)
	}
	if server.WriteTimeout <= bondSearchTimeout || server.WriteTimeout > time.Minute {
		t.Fatalf("write timeout %v must exceed search timeout %v and remain bounded", server.WriteTimeout, bondSearchTimeout)
	}
}
