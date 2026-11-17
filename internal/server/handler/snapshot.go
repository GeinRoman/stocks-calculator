package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"time"

	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/servererrors"
)

const snapshotMaxBytes = 32 << 20

// Snapshot version 1 requires every field, including false, zero and empty arrays.
func snapshotFieldsPresent(raw json.RawMessage, typ reflect.Type) bool {
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return false
	}
	if typ == reflect.TypeOf(time.Time{}) {
		return true
	}
	switch typ.Kind() {
	case reflect.Struct:
		var fields map[string]json.RawMessage
		if json.Unmarshal(raw, &fields) != nil || fields == nil {
			return false
		}
		for i := 0; i < typ.NumField(); i++ {
			f := typ.Field(i)
			name := f.Tag.Get("json")
			value, ok := fields[name]
			if !ok || !snapshotFieldsPresent(value, f.Type) {
				return false
			}
		}
	case reflect.Slice:
		var elements []json.RawMessage
		if json.Unmarshal(raw, &elements) != nil || elements == nil {
			return false
		}
		for _, e := range elements {
			if !snapshotFieldsPresent(e, typ.Elem()) {
				return false
			}
		}
	}
	return true
}

// @Summary Download a snapshot of all user portfolio data
// @Tags snapshot
// @Produce json
// @Success 200 {object} model.Snapshot
// @Failure 401 {string} string
// @Failure 500
// @Security BearerAuth
// @Router /getsnapshot [get]
func (h *handler) getSnapshot(w http.ResponseWriter, r *http.Request) {
	user, e := extractUserId(r)
	if e != nil {
		http.Error(w, e.msg, e.code)
		return
	}
	snapshot, err := h.app.GetSnapshot(r.Context(), user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	data, err := json.Marshal(snapshot)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", `attachment; filename="snapshot.json"`)
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(data)
}

// @Summary Replace all user portfolio data with a snapshot
// @Description Destructively replaces all profiles and portfolio data for the authenticated user. Maximum body size is 32 MiB.
// @Tags snapshot
// @Accept json
// @Param snapshot body model.Snapshot true "Snapshot"
// @Success 200
// @Failure 400 {string} string
// @Failure 401 {string} string
// @Failure 413 {string} string
// @Failure 500
// @Security BearerAuth
// @Router /installsnapshot [post]
func (h *handler) installSnapshot(w http.ResponseWriter, r *http.Request) {
	user, e := extractUserId(r)
	if e != nil {
		http.Error(w, e.msg, e.code)
		return
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, snapshotMaxBytes))
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			http.Error(w, "snapshot too large", http.StatusRequestEntityTooLarge)
		} else {
			http.Error(w, "failed to read snapshot", http.StatusBadRequest)
		}
		return
	}
	var snapshot model.Snapshot
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if !snapshotFieldsPresent(body, reflect.TypeOf(snapshot)) || decoder.Decode(&snapshot) != nil {
		http.Error(w, "invalid snapshot", http.StatusBadRequest)
		return
	}
	if err = h.app.InstallSnapshot(r.Context(), user, snapshot); errors.Is(err, servererrors.ErrInvalidSnapshot) {
		http.Error(w, "invalid snapshot", http.StatusBadRequest)
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
