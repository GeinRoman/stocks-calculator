package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

const userIdContextKey = "userId"

type httpError struct {
	msg  string
	code int
}

func (err *httpError) Error() string {
	return err.msg
}

func NewHttpError(msg string, code int) *httpError {
	return &httpError{msg: msg, code: code}
}

func writeAndMarshal[T any](w http.ResponseWriter, data T) {
	response, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(response)
}

func readBody[T any](r *http.Request) (*T, *httpError) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, NewHttpError("failed to read request body", http.StatusInternalServerError)
	}
	if len(body) == 0 {
		return nil, NewHttpError("request body is missing", http.StatusBadRequest)
	}

	var data T
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, NewHttpError("request body has incorrect format", http.StatusBadRequest)
	}
	return &data, nil
}

func extractUserId(ctx context.Context) (int, *httpError) {
	id, ok := ctx.Value(userIdContextKey).(int)
	if !ok {
		return 0, NewHttpError("Failed to acquire user id", http.StatusInternalServerError)
	}
	return id, nil
}
