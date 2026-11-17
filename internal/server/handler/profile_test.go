package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/servererrors"

	"github.com/stretchr/testify/require"
)

type renameProfileTestApp struct {
	App
	called bool
	user   int
	body   model.RenameProfileBody
	err    error
}

func (a *renameProfileTestApp) RenameProfile(_ context.Context, body model.RenameProfileBody, user int) error {
	a.called, a.user, a.body = true, user, body
	return a.err
}

func TestRenameProfile(t *testing.T) {
	for _, tc := range []struct {
		name, body, token string
		err               error
		status            int
		called            bool
	}{
		{"success", `{"name":"old","new_name":"new"}`, "alice", nil, http.StatusOK, true},
		{"same name", `{"name":"old","new_name":"old"}`, "alice", nil, http.StatusOK, true},
		{"missing old", `{"new_name":"new"}`, "alice", nil, http.StatusBadRequest, false},
		{"missing new", `{"name":"old"}`, "alice", nil, http.StatusBadRequest, false},
		{"invalid json", `{`, "alice", nil, http.StatusBadRequest, false},
		{"unauthorized", `{"name":"old","new_name":"new"}`, "", nil, http.StatusUnauthorized, false},
		{"not found", `{"name":"old","new_name":"new"}`, "alice", servererrors.ErrProfileNotFound, http.StatusNotFound, true},
		{"conflict", `{"name":"old","new_name":"new"}`, "alice", servererrors.ErrProfileExists, http.StatusConflict, true},
		{"failure", `{"name":"old","new_name":"new"}`, "alice", errors.New("failure"), http.StatusInternalServerError, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			app := &renameProfileTestApp{err: tc.err}
			server := NewHttpServer(app, bondTestTokens{}, 0)
			w := bondRequest(server.Handler, "PATCH", "/renameprofile", tc.body, tc.token)
			require.Equal(t, tc.status, w.Code)
			require.Equal(t, tc.called, app.called)
			if tc.called {
				require.Equal(t, 11, app.user)
				require.Equal(t, "old", app.body.Name)
				require.NotEmpty(t, app.body.NewName)
			}
		})
	}
}
